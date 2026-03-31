当向日志追加记录时，会希望此记录要么完全写入，要么完全不写入。

当前的 KV 内存存储结构，已经使用了 fsync() 系统调用来同步文件，因此只有最后一条记录可能会收到断电影响。

使用 checksum 校验和，其本质是一种哈希值，不同的数据对应不同的校验和。使用标准库的 crc32.ChecksumIEEE() 来计算日志记录的校验和，并将其附加到记录之前。当读取记录时，会重新计算读取数据的校验和并与存储的校验和进行比对。

```
|  crc32  | key size | val size | deleted | key data | val data |
| 4 bytes | 4 bytes  | 4 bytes  | 1 byte  |   ...    |   ...    |
```

kv_entry.go 修改 Entry.Encode() 和 Entry.Decode() 方法，编码时添加校验和，解码时比较校验和。

- func (ent *Entry) Encode() []byte
- func (ent *Entry) Decode(r io.Reader) error

Entry.Decode() 方法应返回适当的错误：

- ErrBadSum - 校验和不匹配。
- io.EOF - 已到达文件末尾。
- io.ErrUnexpectedEOF - 文件意外结束。

在 kv_entry.go 中定义 ErrBadSum 错误：

```go
var ErrBadSum = errors.New("bad checksum")
```

