package valse

import "github.com/valyala/fasthttp"

// Fasthttp request handler to valse.RequestHandler
func rWrapper(handler fasthttp.RequestHandler) RequestHandler {
	return func(ctx *Context) error {
		handler(ctx.RequestCtx)
		return nil
	}
}

func mWrapper(r RequestHandler) MiddlewareHandler {
	return func(next RequestHandler) RequestHandler {
		return r
	}
}

func cWrapper(fn func(ctx *Context, next RequestHandler) error) MiddlewareHandler {
	return func(next RequestHandler) RequestHandler {
		return func(ctx *Context) error {
			return fn(ctx, next)
		}
	}
}
