为了保证内存存储的持久性，需要使用日志记录每一次的操作情况。

仅追加日志：在文件末尾追加条目，不修改或删除现有条目。

|      | operation | state        |
| :--- | :-------- | ------------ |
| 0    |           | {}           |
| 1    | set k1=x  | {k1=x}       |
| 2    | set k2=y  | {k1=x, k2=y} |
| 3    | set k1=z  | {k1=z, k2=y} |
| 4    | del k2    | {k1=z}       |

数据库启动时，会读取记录的日志，并将数据库状态设置为 {k1=z}。

后续随着日志增大到一定大小，会被合并到主数据结构（LSM 树或 B+ 树）中。

日志包含两种类型的记录：更新和删除。因此，可以在 kv_entry.go 文件中更新 Entry 数据结构，添加删除标识。当 Entry 为删除日志时，其 deleted = true，且 val 为 nil。

```go
type Entry struct {
	key     []byte
	val     []byte
	deleted bool
}
```

随后序列化格式变为：

```
| key size | val size | deleted | key data | val data |
| 4 bytes  | 4 bytes  | 1 byte  |   ...    |   ...    |
```

修改 kv_entry.go 中的两个序列化和反序列化函数：

- func (ent *Entry) Encode() []byte
- func (ent *Entry) Decode(r io.Reader) error

storage 目录下新增 log.go 文件，定义 Log 结构体，用于记录、存储日志等。

```
storage         
├─ kv.go        
├─ kv_entry.go  
├─ kv_test.go   
└─ log.go       
```

```go
type Log struct {
    FileName string
    fp       *os.File
}

func (log *Log) Open() (err error) {
	log.fp, err = os.OpenFile(log.FileName, os.O_RDWR|os.O_CREATE, 0o644)
	return err
}

func (log *Log) Close() error {
	return log.fp.Close()
}
```

实现 Log.Write() 和 Log.Read() 方法：

- func (log *Log) Write(ent *Entry) error
- func (log *Log) Read(ent *Entry) (eof bool, err error)

File 提供了 Write() 方法用于写入数据，同时 File 实现了 io.Reader 接口，使用 io.EOF 检测文件结束。

最后在 kv.go 的 KV 结构体中添加 log 属性，给定 KV.Close() 方法，要求修改 KV.Open() 方法，使其在初始化数据库后，用日志还原数据库状态。同时修改 KV.Set()、KV.Del() 方法，在方法中使用 log 记录日志。

```go
type KV struct {
	log Log
	mem map[string][]byte
}

func (kv *KV) Close() error {
	return kv.log.Close()
}
```

- func (kv *KV) Open() error
- func (kv *KV) Set(key []byte, val []byte) (updated bool, err error)
- func (kv *KV) Del(key []byte) (deleted bool, err error)

优先执行记录日志操作，然后再对内存数据库进行操作。
