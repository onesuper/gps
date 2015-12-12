# Golang Parser for SQL


Inspired by Rob Pike's [Lexical Scanning in Go](http://cuddle.googlecode.com/hg/talk/lex.html#title-slide) and Project [sql-parser](https://github.com/forward/sql-parser).




## Lexer


The lexer takes a SQL query as input and an alias for it. And each call to `Tokenize()` will get a token from the SQL-query. The EOF- and Error-typed token will be associated with an error.

For example:

```
lexer := gps.NewLexer("lex", "select * from `table` where `a` = xyz")
for {
	tok, err := lexer.Tokenize()
	fmt.Println(tok.Type, tok.String())
	if err != nil {
		break
	}
}

11	"select"
2	"*"
13	"from"
9	"`table`"
14	"where"
9	"`a`"
5	"="
0	ERROR: lex: keyword doesn't exsit: xyz
```


## Parser