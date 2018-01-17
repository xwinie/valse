package valse

import (
	"fmt"
	"testing"
)

func TestServer(t *testing.T) {

	s := New()

	s.Use(func(next RequestHandler) RequestHandler {
		return func(ctx *Context) error {
			fmt.Printf("Middleware 1 %v\n", ctx)
			if err := next(ctx); err != nil {
				return err
			}
			fmt.Println("Middleware 1 after")
			return nil
		}
	})

	s.Use(func(next RequestHandler) RequestHandler {
		return func(ctx *Context) error {
			fmt.Printf("Middleware 2 %v\n", ctx)
			return next(ctx)
		}
	})

	s.Get("/s", func(ctx *Context) error {
		return ctx.JSON(string(ctx.RequestURI()))
	})

	s.Listen(":3000")

}
