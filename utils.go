package valse

import (
	"errors"
	"net/http"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

func toMiddlewareHandler(handler interface{}) (MiddlewareHandler, error) {
	switch h := handler.(type) {
	case func(*Context) error:
		return mWrapper(h), nil
	case func(*fasthttp.RequestCtx):
		return mWrapper(rWrapper(h)), nil
	case func(RequestHandler) RequestHandler:
		return h, nil
	case func(ctx *Context, next RequestHandler) error:
		return cWrapper(h), nil
	case MiddlewareHandler:
		return h, nil

	default:
		return nil, errors.New("middleware is of wrong type")
	}
}

func compose(handlers []interface{}) (RequestHandler, error) {
	last := handlers[len(handlers)-1]
	var routeHandler func(ctx *Context) error
	if fn, ok := last.(func(ctx *Context) error); ok {
		routeHandler = fn
	} else if fn, ok := last.(func(*fasthttp.RequestCtx)); ok {
		routeHandler = rWrapper(fn)
	} else if fn, ok := last.(fasthttp.RequestHandler); ok {
		routeHandler = rWrapper(fn)
	} else if fn, ok := last.(http.Handler); ok {
		routeHandler = rWrapper(fasthttpadaptor.NewFastHTTPHandler(fn))
	} else if fn, ok := last.(http.HandlerFunc); ok {
		routeHandler = rWrapper(fasthttpadaptor.NewFastHTTPHandlerFunc(fn))
	} else if fn, ok := last.(ValseHTTPHandler); ok {
		routeHandler = fn.ServeHTTP
	} else {
		return nil, errors.New("The last handle must be a RequestHandler or a fasthttp Handler")
	}

	var mHandlers []MiddlewareHandler
	for _, h := range handlers[:len(handlers)-1] {
		hh, err := toMiddlewareHandler(h)
		if err != nil {
			return nil, err
		}
		mHandlers = append(mHandlers, hh)
	}

	for i := len(mHandlers) - 1; i >= 0; i-- {
		routeHandler = mHandlers[i](routeHandler)
	}

	// Now compose
	return routeHandler, nil
}

type Route struct {
	Method  string
	Path    string
	Handler RequestHandler
}
