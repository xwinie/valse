package valse

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/gavv/httpexpect"
	"github.com/valyala/fasthttp"
	. "github.com/xwinie/valse"
)

// func TestServer(t *testing.T) {

// 	s := New()

// 	s.Use(func(next RequestHandler) RequestHandler {
// 		return func(ctx *Context) error {
// 			fmt.Printf("Middleware 1 %v\n", ctx)
// 			if err := next(ctx); err != nil {
// 				return err
// 			}
// 			fmt.Println("Middleware 1 after")
// 			return nil
// 		}
// 	})

// 	s.Use(func(next RequestHandler) RequestHandler {
// 		return func(ctx *Context) error {
// 			fmt.Printf("Middleware 2 %v\n", ctx)
// 			return next(ctx)
// 		}
// 	})

// 	s.Get("/s", func(ctx *Context) error {
// 		return ctx.JSON(string(bytes.TrimLeft(ctx.RequestURI(), "/")))
// 	})

// 	s.Listen(":3000")

// }

func TestHttpServer(t *testing.T) {
	s := New()
	s.Get("/s", func(ctx *Context) error {
		return ctx.JSON(string(bytes.TrimLeft(ctx.RequestURI(), "/")))
	})
	s.Get("/n", func(ctx *Context) error {
		return ctx.JSON(string(bytes.TrimLeft(ctx.RequestURI(), "/")))
	})
	e := fastHTTPTester(t, s.GetHandler())
	e.GET("/s").Expect().
		Status(200).
		Text().Equal("pong")
	e.GET("/n").Expect().
		Status(200).
		Text().Equal("pong")
}

// fastHTTPTester returns a new Expect instance to test FastHTTPHandler().
func fastHTTPTester(t *testing.T, h fasthttp.RequestHandler) *httpexpect.Expect {
	return httpexpect.WithConfig(httpexpect.Config{
		// Pass requests directly to FastHTTPHandler.
		Client: &http.Client{
			Transport: httpexpect.NewFastBinder(h),
			Jar:       httpexpect.NewJar(),
		},
		// Report errors using testify.
		Reporter: httpexpect.NewAssertReporter(t),
	})
}
