数据库的实现从 KV 存储系统开始，首先创建基于内存的 KV 存储系统，可以直接复用 map 类型。

创建 storage 文件夹，以及 kv.go、kv_test.go 文件。

```
storage        
├─ kv.go       
└─ kv_test.go
```

kv.go 给定 KV 结构、KV.Open() 和 KV.Close() 方法。

```go
type KV struct {
	mem map[string][]byte
}

func (kv *KV) Open() error {
	kv.mem = make(map[string][]byte)
	return nil
}

func (kv *KV) Close() error {
	return nil
}
```

实现 KV.Get()、KV.Set()、KV.Del() 方法：

- func (kv *KV) Get(key []byte) (val []byte, ok bool, err error)
- func (kv *KV) Set(key []byte, val []byte) (updated bool, err error)
- func (kv *KV) Del(key []byte) (deleted bool, err error)

