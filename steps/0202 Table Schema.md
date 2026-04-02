新建 database/row.go 文件，对一个表进行建模。

``` 
database         
├─ cell.go       
├─ cell_test.go  
├─ row.go        
└─ row_test.go   
```

一个表包含表名、列名、列类型、主键，因此创建 Schema 和 Column 结构体来表示：

```go
type Schema struct {
	Table string
	Cols  []Column
	PKey  []int
}

type Column struct {
	Name string
	Type CellType
}
```

同时，实际保存的行数据是多个 Cell 构成的集合，因此创建 Row 结构体：

```go
type Row []Cell

func (schema *Schema) NewRow() Row {
    return make(Row, len(schema.Cols))
}
```

Schema.NewRow() 返回一个大小合适但未初始化的行。

实现 Row 对应的序列化和反序列化方法：

- func (row Row) EncodeKey(schema *Schema) (key []byte)
- func (row Row) EncodeVal(schema *Schema) (val []byte)
- func (row Row) DecodeKey(schema *Schema, key []byte) (err error)
- func (row Row) DecodeVal(schema *Schema, val []byte) (err error)

其中，对 Row 的主键与非主键列分别进行编码，主键是 KV 存储结构中的 key，其余列是 KV 存储结构中的 val，可以通过主键值寻找到其它列值。一个数据库包含多张表，因此编码后的主键的 key 需要前缀来区分不同表，使用表名作为前缀：

```go
func (row Row) EncodeKey(schema *Schema) (key []byte) {
    key = append([]byte(schema.Table), 0x00)
    
    ...
}
```

在表名后添加一个 0x00 分隔符，以避免因名称前缀导致的键冲突。