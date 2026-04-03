现在可以将 SQL 操作映射到 KV 方法，例如 SELECT 对应 KV.Get()，DELETE -> KV.Del()。但是 KV.Set() 方法当前包括了 INSERT 和 UPDATE 两种模式，实际上表示了 INSERT ... ON DUPLICATE UPDATE。

在 storage/kv.go 中，为 KV 内存存储添加 KV.SetEx() 方法，以此来区分 INSERT 和 UPDATE。首先新建一个 UpdateMode 结构，标识不同的模式。

```go
type UpdateMode int

const (
	ModeUpsert UpdateMode = 0 // insert or update
	ModeInsert UpdateMode = 1 // insert new
	ModeUpdate UpdateMode = 2 // update existing
)
```

- func (kv *KV) SetEx(key []byte, val []byte, mode UpdateMode) (updated bool, err error)
  - ModeInsert：如果键已存在，则不更新并返回 false。
  - ModeUpdate：仅更新现有键。
  - ModeUpsert：与旧的 KV.Set() 方法相同，插入或覆盖。

最后修改 KV.Set() 方法，使其复用 KV.SetEx() 方法：

```go
func (kv *KV) Set(key []byte, val []byte) (updated bool, err error) {
	return kv.SetEx(key, val, ModeUpsert)
}
```

