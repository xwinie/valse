# valse
Lightweight FastHTTP middleware based HTTP server framework
感谢@kildevaeld，这个是fork kildevaeld had project

## Usage

```go

s := valse.New();

// Using verbs: Get, Post, Put, Delete
s.Get("/", func(ctx *valse.Context) error {
  return ctx.Text("Hello, World")
})
// With middlewares
s.Post("/post", func(ctx *valse.Context, next valse.RequestHandler) error {
  fmt.Println("before")
  next(ctx)
  fmt.Println("after")
}, func (ctx *valse.Context) error {
  return ctx.JSON(dict.Map{"Hello": "World"})
})




```
