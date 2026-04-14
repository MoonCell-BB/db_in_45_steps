当前的 KVIterator 只支持开放式范围查询，即 key >= 123。先新增一个 RangedKVIter 结构体，其支持闭合式范围查询，且使用一个 desc 标志，其为 false 时查询 [start, stop]，其为 true 时查询 [stop, start]。

storage/kv.go 新增 RangedKVIter 结构体，其是对 KVIterator 结构体的封装：

```go
type RangedKVIter struct {
	iter KVIterator
	stop []byte
	desc bool
}
```

- func (kv \*KV) Range(start, stop []byte, desc bool) (\*RangedKVIter, error)
- func (iter *RangedKVIter) Valid() bool
- func (iter *RangedKVIter) Key() []byte
- func (iter *RangedKVIter) Val() []byte
- func (iter *RangedKVIter) Next() error

RangedKVIter.Valid() 方法会判定 stop 与当前 key 的关系，在适当时终止迭代。

现在为 database 中的数据库添加一个支持范围的 Range() 方法，需要考虑到主键或索引可以由多列组成，而查询可能只使用前缀：

- (a, b, c) <= (123, 4, 5) 为 stop 提供了完整的键，此时可以直接使用 RangedKVIter。
- (a, b) <= (123, 4) 使用了不完整的键，实际上等同于提供了 (123, 4, +∞)。
- a <= 123 也使用了不完整的键，实际上等同于提供了 (123, +∞, +∞)。

此外，若支持 (a, b) < (123, 4) 的形式，则其等同于提供了 (a, b, c) <= (123, 4, -∞) 的完整键。

在 database/row.go 中添加 EncodeKeyPrefix() 函数，用于将提供的不完整 prefix []Cell 填充为完整键。此外，该函数中还有一个 positive 标识，用于区分填充部分为 +∞ 还是 -∞。

- "" 表示 −∞，这要求序列化数据绝不能是 ""：
  - int64：始终会填写 8 字节。
  - string：末尾会添加 0x00 作为后缀。
- \xff 表示 +∞，这要求序列化数据绝不能以 "\xff" 开头，可通过为每列前置 1 字节来确保。

```go
func (row Row) EncodeKey(schema *Schema) (key []byte) {
	if len(row) != len(schema.Cols) {
		panic("mismatch between row data and schema")
	}

	key = append([]byte(schema.Table), 0x00)

	for _, idx := range schema.PKey {
		cell := row[idx]
		if cell.Type != schema.Cols[idx].Type {
			panic("cell type mismatch")
		}

		key = append(key, byte(cell.Type))
		key = cell.EncodeKey(key)
	}

	return append(key, 0x00)
}
```

原始的 Row.EncodeKey() 方法序列化后的数据格式为：table_name | 0x00 | 123 | "abc"。现在在序列化每个 cell 时，会添加一个单独的 cell.Type 字节，且在末尾添加一个额外的 0x00，格式为：

table_name | 0x00 | 0x01(int) | 123 | 0x02 | "abc" | 0x00。

此时编写函数 EncodeKeyPrefix()，若只提供了 123，则根据 positive 标志将其转化为不同序列结构：

- positive 为 true，则表示序列化为 +∞，序列化为 table_name | 0x00 | 0x01(int) | 123 | 0xff。
- positive 为 false，则表示序列化为 -∞，序列化为 table_name | 0x00 | 0x01(int) | 123。

使用此序列化格式，即使不提供 cell，也会序列化为 table_name | 0x00 或 table_name | 0x00 | 0xff 来分别表示 -∞ 和 +∞。

提供所有的 cell，会将其序列化为 table_name | 0x00 | ... | "" 和 table_name | 0x00 | ... | 0xff 来分别表示 -∞ 和 +∞。

- func EncodeKeyPrefix(schema *Schema, prefix []Cell, positive bool) []byte

```go
func EncodeKeyPrefix(schema *Schema, prefix []Cell, positive bool) []byte {
	if len(prefix) != len(schema.Cols) {
		panic("mismatch between row data and schema")
	}

	key := append([]byte(schema.Table), 0x00)
	for idx, cell := range prefix {
		if cell.Type != schema.Cols[schema.PKey[idx]].Type {
			panic("cell type mismatch")
		}

		key = append(key, byte(cell.Type))
		key = cell.EncodeKey(key)
	}

	if positive {
		key = append(key, 0xff)
	}

	return key
}
```

同步修改 Row.DecodeKey() 方法，在解析时需要跳过新增类型字段：

- func (row Row) DecodeKey(schema *Schema, key []byte) (err error)

```go
func (row Row) DecodeKey(schema *Schema, key []byte) (err error) {
	...

	for _, idx := range schema.PKey {
		col := schema.Cols[idx]
		if len(key) == 0 {
			return ErrDataLen
		}
		
		if CellType(key[0]) != col.Type {
			return errors.New("cell type mismatch")
		}
		key = key[1:]

		row[idx] = Cell{Type: col.Type}

		if key, err = row[idx].DecodeKey(key); err != nil {
			return err
		}
	}

	if len(key) != 1 || key[0] != 0x00 {
		return errors.New("trailing garbage")
	}

	return nil
}
```

database 新增 operator.go，用于定义操作符枚举值以及定义 RangeReq，用于存放比较表达式：

```
database          
├─ cell.go        
├─ cell_test.go   
├─ operator.go    
├─ row.go         
├─ row_test.go    
├─ row_utils.go   
├─ table.go       
└─ table_test.go  
```

```go
type ExprOp uint8

const (
	OP_LE ExprOp = 12 // <=
	OP_GE ExprOp = 13 // >=
	OP_LT ExprOp = 14 // <
	OP_GT ExprOp = 15 // >
)

type RangeReq struct {
	StartCmp ExprOp
	StopCmp  ExprOp
	Start    []Cell
	Stop     []Cell
}

func IsDescending(op ExprOp) bool {
	switch op {
	case OP_LE, OP_LT:
		return true
	case OP_GE, OP_GT:
		return false
	default:
		panic("unreachable")
	}
}

func SuffixPositive(op ExprOp) bool {
	switch op {
	case OP_LE, OP_GT:
		return true
	case OP_GE, OP_LT:
		return false
	default:
		panic("unreachable")
	}
}
```

| Prefix           | Full                    |
| :--------------- | :---------------------- |
| (a, b) < (1, 2)  | (a, b, c) <= (1, 2, −∞) |
| (a, b) <= (1, 2) | (a, b, c) <= (1, 2, +∞) |
| (a, b) > (1, 2)  | (a, b, c) >= (1, 2, +∞) |
| (a, b) >= (1, 2) | (a, b, c) >= (1, 2, −∞) |

例如 b < 1 的表达式可以将其返回表示为 (−∞, 1)，遍历方向由 StartCmp 决定：>= 和 > 表示升序，否则为降序。

database/table.go 中定义的 RowIterator 结构体，其内部使用新的 RangedKVIter，同时 decodeKVIter() 函数使用新的 RangedKVIter 结构体：

```go
type RowIterator struct {
	schema *Schema
	iter   *storage.RangedKVIter
	valid  bool
	row    Row
}

func decodeKVIter(schema *Schema, iter *storage.RangedKVIter, row Row) (bool, error) {
	...   
}
```

提供 DB.Range() 方法，用于范围遍历数据库数据：

- func (db \*DB) Range(schema \*Schema, req \*RangeReq) (\*RowIterator, error)

使用新的 DB.Range() 方法，修改之前的 DB.Seek() 方法，等价于使用 StartCmp = OP_GE 的方式；

```go
func (db *DB) Seek(schema *Schema, row Row) (*RowIterator, error) {
	start := make([]Cell, len(schema.PKey))
	for i, idx := range schema.PKey {
		if row[idx].Type != schema.Cols[idx].Type {
			panic("cell type mismatch")
		}

		start[i] = row[idx]
	}

	return db.Range(schema, &RangeReq{
		StartCmp: OP_GE,
		StopCmp:  OP_LE,
		Start:    start,
		Stop:     nil,
	})
}
```

