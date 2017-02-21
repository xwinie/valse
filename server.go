package valse

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/buaazp/fasthttprouter"
	"github.com/kildevaeld/strong"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

type RequestHandler func(*Context) error
type MiddlewareHandler func(next RequestHandler) RequestHandler
type ValseHTTPHandler interface {
	ServeHTTP(*Context) error
}

func notFoundOrErr(ctx *Context, err error) error {
	if ctx.Response.StatusCode() == strong.StatusNotFound || err == nil {
		return nil
	}
	status := http.StatusInternalServerError
	if e, ok := err.(*strong.HttpError); ok {
		ctx.Error(e.Message(), e.Code())
		return nil
	}

	ctx.Error(err.Error(), status)

	return nil
}

type Config struct {
	// Handler for processing incoming requests.
	//Handler RequestHandler

	// Server name for sending in response headers.
	//
	// Default server name is used if left blank.
	Name string

	// The maximum number of concurrent connections the server may serve.
	//
	// DefaultConcurrency is used if not set.
	Concurrency int

	// Whether to disable keep-alive connections.
	//
	// The server will close all the incoming connections after sending
	// the first response to client if this option is set to true.
	//
	// By default keep-alive connections are enabled.
	DisableKeepalive bool

	// Per-connection buffer size for requests' reading.
	// This also limits the maximum header size.
	//
	// Increase this buffer if your clients send multi-KB RequestURIs
	// and/or multi-KB headers (for example, BIG cookies).
	//
	// Default buffer size is used if not set.
	ReadBufferSize int

	// Per-connection buffer size for responses' writing.
	//
	// Default buffer size is used if not set.
	WriteBufferSize int

	// Maximum duration for reading the full request (including body).
	//
	// This also limits the maximum duration for idle keep-alive
	// connections.
	//
	// By default request read timeout is unlimited.
	ReadTimeout time.Duration

	// Maximum duration for writing the full response (including body).
	//
	// By default response write timeout is unlimited.
	WriteTimeout time.Duration

	// Maximum number of concurrent client connections allowed per IP.
	//
	// By default unlimited number of concurrent connections
	// may be established to the server from a single IP address.
	MaxConnsPerIP int

	// Maximum number of requests served per connection.
	//
	// The server closes connection after the last request.
	// 'Connection: close' header is added to the last response.
	//
	// By default unlimited number of requests may be served per connection.
	MaxRequestsPerConn int

	// Maximum keep-alive connection lifetime.
	//
	// The server closes keep-alive connection after its' lifetime
	// expiration.
	//
	// See also ReadTimeout for limiting the duration of idle keep-alive
	// connections.
	//
	// By default keep-alive connection lifetime is unlimited.
	MaxKeepaliveDuration time.Duration

	// Maximum request body size.
	//
	// The server rejects requests with bodies exceeding this limit.
	//
	// Request body size is limited by DefaultMaxRequestBodySize by default.
	MaxRequestBodySize int

	// Aggressively reduces memory usage at the cost of higher CPU usage
	// if set to true.
	//
	// Try enabling this option only if the server consumes too much memory
	// serving mostly idle keep-alive connections. This may reduce memory
	// usage by more than 50%.
	//
	// Aggressive memory usage reduction is disabled by default.
	ReduceMemoryUsage bool

	// Rejects all non-GET requests if set to true.
	//
	// This option is useful as anti-DoS protection for servers
	// accepting only GET requests. The request size is limited
	// by ReadBufferSize if GetOnly is set.
	//
	// Server accepts all the requests by default.
	GetOnly bool

	// Logs all errors, including the most frequent
	// 'connection reset by peer', 'broken pipe' and 'connection timeout'
	// errors. Such errors are common in production serving real-world
	// clients.
	//
	// By default the most frequent errors such as
	// 'connection reset by peer', 'broken pipe' and 'connection timeout'
	// are suppressed in order to limit output log traffic.
	LogAllErrors bool

	// Header names are passed as-is without normalization
	// if this option is set.
	//
	// Disabled header names' normalization may be useful only for proxying
	// incoming requests to other servers expecting case-sensitive
	// header names. See https://github.com/valyala/fasthttp/issues/57
	// for details.
	//
	// By default request and response header names are normalized, i.e.
	// The first letter and the first letters following dashes
	// are uppercased, while all the other letters are lowercased.
	// Examples:
	//
	//     * HOST -> Host
	//     * content-type -> Content-Type
	//     * cONTENT-lenGTH -> Content-Length
	DisableHeaderNamesNormalizing bool

	// Logger, which is used by RequestCtx.Logger().
	//
	// By default standard logger from log package is used.
	Logger Logger
}

type Server struct {
	noCopy
	s       *fasthttp.Server
	r       *fasthttprouter.Router
	running bool
	m       []MiddlewareHandler
	p       sync.Pool
}

func (s *Server) Use(handlers ...interface{}) *Server {
	if s.running {
		panic("cannot add middleware when running.")
	}

	for _, handler := range handlers {
		switch h := handler.(type) {
		case func(*Context) error:
			s.m = append(s.m, mWrapper(h))
		case func(*fasthttp.RequestCtx):
			s.m = append(s.m, mWrapper(rWrapper(h)))
		case func(RequestHandler) RequestHandler:
			s.m = append(s.m, h)
		case func(ctx *Context, next RequestHandler) error:
			s.m = append(s.m, cWrapper(h))
		case MiddlewareHandler:
			s.m = append(s.m, h)
		default:
			panic(fmt.Sprintf("middleware is of wrong type %t", handler))
		}
	}

	return s
}

func (s *Server) Get(path string, handlers ...interface{}) *Server {
	return s.Route(strong.GET, path, handlers...)
}

func (s *Server) Post(path string, handlers ...interface{}) *Server {
	return s.Route(strong.POST, path, handlers...)
}

func (s *Server) Put(path string, handlers ...interface{}) *Server {
	return s.Route(strong.PUT, path, handlers...)
}

func (s *Server) Delete(path string, handlers ...interface{}) *Server {
	return s.Route(strong.DELETE, path, handlers...)
}

func (s *Server) Head(path string, handlers ...interface{}) *Server {
	return s.Route(strong.HEAD, path, handlers...)
}

func (s *Server) Options(path string, handlers ...interface{}) *Server {
	return s.Route(strong.OPTIONS, path, handlers...)
}

func (s *Server) Route(method, path string, handlers ...interface{}) *Server {
	if len(handlers) == 0 {
		return s
	}

	handler, err := s.compose(handlers)

	if err != nil {
		panic(err)
	}

	s.r.Handle(method, path, s.handleRequest(handler))

	return s
}

func (s *Server) toMiddlewareHandler(handler interface{}) (MiddlewareHandler, error) {
	switch h := handler.(type) {
	case func(*Context) error:
		return mWrapper(h), nil
	case func(*fasthttp.RequestCtx):
		return mWrapper(rWrapper(h)), nil
	case func(RequestHandler) RequestHandler:
		return h, nil
	case func(ctx *Context, next RequestHandler) error:
		return cWrapper(h), nil

	default:
		return nil, errors.New("middleware is of wrong type")
	}
}

func (s *Server) Listen(address string) error {

	if s.running {
		return errors.New("Already running")
	}
	s.running = true

	handlers := rWrapper(s.r.Handler)
	for i := len(s.m) - 1; i >= 0; i-- {
		handlers = s.m[i](handlers)
	}

	s.s.Handler = s.handleRequest(handlers)

	return s.s.ListenAndServe(address)
}

func (s *Server) compose(handlers []interface{}) (RequestHandler, error) {
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
		hh, err := s.toMiddlewareHandler(h)
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

func (s *Server) handleRequest(handler RequestHandler) fasthttp.RequestHandler {
	return func(requestCtx *fasthttp.RequestCtx) {
		ctx := s.p.Get().(*Context)
		ctx.RequestCtx = requestCtx
		defer func() { s.p.Put(ctx.reset()) }()
		if err := handler(ctx); err != nil {
			notFoundOrErr(ctx, err)
		}
	}
}

func New() *Server {
	return newWithServer(&fasthttp.Server{})
}

func NewWithConfig(config Config) *Server {

	s := &fasthttp.Server{
		Concurrency:                   config.Concurrency,
		Name:                          config.Name,
		DisableKeepalive:              config.DisableKeepalive,
		ReadBufferSize:                config.ReadBufferSize,
		WriteBufferSize:               config.WriteBufferSize,
		WriteTimeout:                  config.WriteTimeout,
		MaxRequestsPerConn:            config.MaxRequestsPerConn,
		MaxKeepaliveDuration:          config.MaxKeepaliveDuration,
		MaxRequestBodySize:            config.MaxRequestBodySize,
		ReduceMemoryUsage:             config.ReduceMemoryUsage,
		GetOnly:                       config.GetOnly,
		LogAllErrors:                  config.LogAllErrors,
		DisableHeaderNamesNormalizing: config.DisableHeaderNamesNormalizing,
		Logger: config.Logger,
	}

	return newWithServer(s)
}

func newWithServer(server *fasthttp.Server) *Server {
	return &Server{
		s: server,
		r: fasthttprouter.New(),
		p: sync.Pool{
			New: func() interface{} {
				return &Context{}
			},
		},
	}
}
