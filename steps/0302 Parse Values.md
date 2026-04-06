由于本数据库只实现了 int64 和 string 两种数据类型，因此只需解析 1、'hello' 这两种形式的值。

- int64 - 以 - 或 + 开头，后跟数字。
- string - 由 ' ' 或 " " 括起来。

给定 Parser.parseValue() 方法，其会根据后续单词的首字符来判断调用解析字符串还是调用解析数字的方法。

```go
func (p *Parser) parseValue(out *database.Cell) error {
	if p.isEnd() {
		return errors.New("expect value")
	}

	ch := p.buf[p.pos]
	if ch == '"' || ch == '\'' {
		return p.parseString(out)
	} else if IsDigit(ch) || ch == '-' || ch == '+' {
		return p.parseInt(out)
	} else {
		return errors.New("expect value")
	}
}
```

实现 Parser.parseString() 和 Parser.parseInt() 方法，分别用于解析 string 和 int64：

- func (p *Parser) parseInt(out *database.Cell) error
- func (p *Parser) parseString(out *database.Cell) error

Parser.parseInt() 方法中，可以使用 go 内置的 strconv.ParseInt() 方法将 string 类型转化为 int 类型。

Parser.parseString() 方法中，字符串内部的引号必须用反斜杠进行转义，例如 'say:\\'hi\\'' 需要转义为 'say:'hi''。