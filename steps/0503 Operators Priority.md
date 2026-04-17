引入 * 和 / 运算符，它们的优先级高于 + 和 -，因此加减运算必须位于语法树的顶层（根节点）。+ 和 - 的子树可以包含 * 和 / 运算，但 \* 和 / 的子树不能包含 + 或 - 运算，只能包含数值或列名。

在 database/operator.go 中添加乘除法的枚举值：

```go
const (
	OP_ADD ExprOp = 1  // +
	OP_SUB ExprOp = 2  // -
	OP_MUL ExprOp = 3  // *
	OP_DIV ExprOp = 4  // /
	OP_LE  ExprOp = 12 // <=
	OP_GE  ExprOp = 13 // >=
	OP_LT  ExprOp = 14 // <
	OP_GT  ExprOp = 15 // >
)
```

Parser.parseAdd() 方法利用 Parser.parserAtom() 方法来解析数值或列名。现在在其中添加一个额外的 Parser.parseMul() 方法，在解析数值、列名的前提下，还可以解析乘除表达式，形式类似于 Parser.parseAdd()：

- func (p *Parser) parseMul() (any, error)

```go
func (p *Parser) parseMul() (any, error) {
	left, err := p.parseAtom()
	if err != nil {
		return nil, err
	}

	tokens := []string{"*", "/"}
	ops := []database.ExprOp{database.OP_MUL, database.OP_DIV}

	for ok := true; ok; {
		ok = false
		for idx, token := range tokens {
			if !p.tryPunctuation(token) {
				continue
			}

			ok = true
			right, err := p.parseAtom()
			if err != nil {
				return nil, err
			}

			left = &ExprBinOp{
				op:    ops[idx],
				left:  left,
				right: right,
			}

			break
		}
	}

	return left, nil
}
```

修改 Parser.parseAdd() 方法，使其使用解析乘除法的 Parser.parseMul() 方法解析加减法：

```go
func (p *Parser) parseAdd() (any, error) {
	left, err := p.parseMul()
	if err != nil {
		return nil, err
	}

	tokens := []string{"+", "-"}
	ops := []database.ExprOp{database.OP_ADD, database.OP_SUB}

	for ok := true; ok; {
		ok = false
		for idx, token := range tokens {
			if !p.tryPunctuation(token) {
				continue
			}

			ok = true
			right, err := p.parseMul()
			if err != nil {
				return nil, err
			}

			left = &ExprBinOp{
				op:    ops[idx],
				left:  left,
				right: right,
			}

			break
		}
	}

	return left, nil
}
```

综上可以得出，表达式的解析是一层嵌套一层的。例如 a * b + c / 3，parseAtom() 用于解析数值或列名，下一个 parseMul() 在此基础上解析乘除表达式，此时 a * b 可被看作 (a * b) = d，c / 3 可被看作 (c / 3) = e，此时表达式可以写为 d + e，最后使用 parseAdd() 来解析加减表达式。

因此，可以创建一个统一的 Parser.ParseExpr() 方法用于解析表达式，在目前调用的是 parseAdd() 方法，后续还有其它表达式优先级，可以灵活调整调用策略。

- func (p *Parser) ParseExpr() (any, error)

```go
func (p *Parser) ParseExpr() (any, error) {
	return p.parseAdd()
}
```

之后，添加括号来改变运算顺序。括号可被视为最高优先级的运算，因此需要在最深层处调用的 Parser.parseAtom() 方法进行处理。当遇到括号时，调用顶层函数 Parser.ParseExpr() 来解析完整表达式。

- func (p *Parser) parseAtom() (expr any, err error)

```go
func (p *Parser) parseAtom() (expr any, err error) {
	if p.tryPunctuation("(") {
		if expr, err = p.ParseExpr(); err != nil {
			return nil, err
		}

		if !p.tryPunctuation(")") {
			return nil, errors.New("expect )")
		}

		return expr, nil
	}

	...
}
```

