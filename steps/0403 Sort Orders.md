当前是以 bytes.Compare() 进行排序比较的。但是关系型数据具有类型，序列化后的键可能以错误的顺序进行比较。

key 以 table_name + 0x00 作为单行数据的前缀标识，后跟主键列的数据构成的。其中 int64 以 8 bytes 格式进行编码，string 以 4 bytes 的方式存储长度，后面跟随编码的数据。 

```
| length  | data |
| 4 bytes | ...  |
```

一种解决方案是在比较前反序列化键为 row，然后再按数据类型进行比较。但这使得键值对依赖于模式，增加了紧密耦合。因此，一些数据库采用另一种方式：设计一种保持排序顺序的序列化格式。

database/cell.go 将旧的 Cell.Encode() 重命名为 Cell.EncodeVal()、Cell.Decode() 重命名为 Cell.DecodeVal()。新增方法 Cell.EncodeKey() 和 Cell.DecodeKey()，以处理键值对。

- func (cell *Cell) EncodeKey(toAppend []byte) []byte
- func (cell *Cell) DecodeKey(data []byte) (rest []byte, err error)

对于 int64 类型，可以将其转化为 uint 并使用大端存储，其最高位保持在最左边，此时会按照从小到大排序。

对于 string 类型，则不能够再使用 length 保存字符串长度，否则就会按照字符串长度大小排序，不符合排序原则。可以使用分隔符来标记字符串的结束：比较两个以零结尾的字符串时，如果一个是另一个的前缀，比较会到达零字节处，较短的字符串会被视为较小。

需要注意，如果字符串数据包含零字节，必须进行转义处理，这里使用 0x01 作为转义字节：

- 0x00 <=> 0x01 0x01
- 0x01 <=> 0x01 0x02

编写 encodeStrKey()、decodeStrKey() 辅助方法：

- func encodeStrKey(toAppend []byte, input []byte) []byte
- func decodeStrKey(data []byte) (rest, str []byte, err error)

编写 Cell.EncodeKey() 和 Cell.EncodeKey() 方法：

- func (cell *Cell) EncodeKey(toAppend []byte) []byte
- func (cell *Cell) DecodeKey(data []byte) (rest []byte, err error)

此外，还需要在 database/row.go 的相关方法中，替换 cell.go 中编写的方法：

- func (row Row) EncodeKey(schema *Schema) (key []byte)
- func (row Row) EncodeVal(schema *Schema) (val []byte)
- func (row Row) DecodeKey(schema *Schema, key []byte) (err error)
- func (row Row) DecodeVal(schema *Schema, val []byte) (err error)



