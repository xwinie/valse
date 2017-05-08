# valse
Lightweight FastHTTP middleware based HTTP server framework


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
