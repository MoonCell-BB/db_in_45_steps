一个数据库可容纳多张表，每张表包含行与列，此数据库实现两种数据类型：int64 与 []byte。

新建 database 目录，在此目录下新建 cell.go 文件，其中定义数据的最小单元。

```
database    
└─ cell.go
```

```go
type CellType uint8

const (
	TypeI64 CellType = 1
	TypeStr CellType = 2
)

type Cell struct {
	Type CellType
	I64  int64
	Str  []byte
}
```

单个 Cell 对象需要实现序列化和反序列化方法：

- func (cell *Cell) Encode(toAppend []byte) []byte
- func (cell *Cell) Decode(data []byte) (rest []byte, err error)

序列化 Cell.Encode() 方法需要从给定的 toAppend 字节切片后，添加本身的序列化字节切片，多个 Cell 值将被连接。反序列化 Cell.Decode() 方法需要从 data 中读取 Cell 对象，并返回剩余数据。

int64 使用 binary.LittleEndian 提供的方法将其序列化为固定 8 长度字节，[]byte 采用长度 + 数据的格式：

```
| length  | data |
| 4 bytes | ...  |
```

