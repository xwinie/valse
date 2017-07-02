//go:generate go-bindata -pkg lua prelude.lua
package lua

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/aarzilli/golua/lua"
	"github.com/kildevaeld/valse"
	"github.com/stevedonovan/luar"
)

func createRequest(ctx *valse.Context) luar.Map {
	return luar.Map{
		"header": luar.Map{
			"get": func(key string) string {
				return string(ctx.Request.Header.Peek(key))
			},
		},
		"url": func() luar.Map {
			url := ctx.Request.URI()
			return luar.Map{
				"query": luar.Map{
					"get": func(key string) string {
						return string(url.QueryArgs().Peek(key))
					},
				},
			}
		},
		"path":   string(ctx.Path()),
		"method": string(ctx.Method()),
		"body": func() string {
			return string(ctx.PostBody())
		},
	}
}

func createResponse(ctx *valse.Context) luar.Map {
	return luar.Map{
		"header": luar.Map{
			"get": func(key string) string {
				return string(ctx.Response.Header.Peek(key))
			},
			"set": func(key, val string) {
				ctx.Response.Header.Set(key, val)
			},
		},
		"write": func(str string) {
			ctx.WriteString(str)
		},
		"setStatus": func(status int) {
			ctx.Response.SetStatusCode(status)
		},
	}
}

type LuaOptions struct {
	Path        string
	StopOnError bool
	Lua         *lua.State
}

func Lua(options LuaOptions) valse.MiddlewareHandler {

	logger := logrus.WithField("prefix", "middlewares:lua")

	L := options.Lua
	if L == nil {
		L = luar.Init()
		L.OpenLibs()
	}

	L.DoString(string(MustAsset("prelude.lua")))

	files, err := ioutil.ReadDir(options.Path)
	if err != nil {
		logger.Fatal(err)
	}

	for _, file := range files {

		ext := filepath.Ext(file.Name())
		if ext == ".lua" && !strings.HasPrefix(file.Name(), "_") {
			logger.Debugf("loading file: %s", file.Name())
			fullPath := filepath.Join(options.Path, file.Name())
			err := L.DoFile(fullPath)
			if err != nil {
				if options.StopOnError {
					logger.Fatal(err)
				}
				logger.WithError(err).Errorf("could not load file: %s", fullPath)
			}
		}
	}

	return func(next valse.RequestHandler) valse.RequestHandler {
		return func(ctx *valse.Context) error {

			req := createRequest(ctx)
			res := createResponse(ctx)

			L.GetGlobal("runMiddlewares")

			luar.GoToLua(L, req)
			luar.GoToLua(L, res)
			L.MustCall(2, 1)
			if !L.ToBoolean(0) {
				return nil
			}

			return next(ctx)
		}
	}
}
