当前已经实现了 Parser.parseAdd() 方法来解析加减表达式，现在可以为 SELECT 和 UPDATE 语句添加计算功能：

```sql
SELECT a + b, c * d FROM t WHERE a = 123;
UPDATE t SET x = x + 1 WHERE a = 123;
```

现在给定一个 expr 表达式，其可以是 string、Cell 和 ExprBinOP 这三种类型，现在需要解析类型，并返回一个统一的 Cell 类型，即最终值是 string 或 int64。

新建 parser/eval.go 文件：

```
parser                  
├─ eval.go              
├─ sql_exec.go          
├─ sql_exec_test.go     
├─ sql_parser.go        
├─ sql_parser_test.go   
└─ sql_parser_utils.go  
```

定义 evalExpr() 函数，其用于解析 expr 表达式并返回 Cell 类型结果：

- func evalExpr(schema \*database.Schema, row database.Row, expr any) (\*database.Cell, error)

如果给定的是 string，可认为其给定的是数据表中的列数据，例如 x + 1 中的 x，则返回 Row 中对应的 Cell；如果给定的是 Cell，则可以直接返回；若给定的是 ExprBinOP 类型，则可以递归调用 evalExpr() 函数进行解析。





