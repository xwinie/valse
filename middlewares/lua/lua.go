//go:generate go-bindata -pkg lua prelude.lua
package lua

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
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
	LuaFactory  func() *lua.State
	WorkQueue   int
}

type File struct {
	Path    string
	Content string
}

func createLua(options LuaOptions, logger logrus.FieldLogger, files []File) *lua.State {
	var L *lua.State

	if options.LuaFactory != nil {
		L = options.LuaFactory()
	} else {
		L = luar.Init()
		L.OpenLibs()
	}

	L.DoString(string(MustAsset("prelude.lua")))

	for _, file := range files {
		logger.Debugf("loading file: %s", file.Path)

		err := L.DoString(file.Content)
		if err != nil {
			if options.StopOnError {
				logger.Fatal(err)
			}
			logger.WithError(err).Errorf("could not load file: %s", file.Path)
		}

	}
	return L
}

func getSortedFiles(path string) ([]string, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var fileNames []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		ext := filepath.Ext(file.Name())
		if ext != ".lua" || strings.HasPrefix(file.Name(), "_") {
			continue
		}

		fileNames = append(fileNames, filepath.Join(path, file.Name()))
	}

	sort.Strings(fileNames)

	luaPath := os.Getenv("LUA_PATH")
	if luaPath != "" {
		luaPath += ";"
	}

	luaPath += path + "/?.lua"

	os.Setenv("LUA_PATH", luaPath)

	return fileNames, nil
}

func Lua(options LuaOptions) valse.MiddlewareHandler {

	logger := logrus.WithField("prefix", "middlewares:lua")

	files, err := getSortedFiles(options.Path)
	if err != nil {
		logger.Fatal(err)
	}
	var out []File
	for _, file := range files {

		bs, err := ioutil.ReadFile(file)
		if err != nil {
			logger.Fatal(err)
		}
		out = append(out, File{
			Path:    file,
			Content: string(bs),
		})
	}

	//var lock sync.Mutex
	wn := options.WorkQueue
	if wn == 0 {
		wn = 5
	}

	ch := make(chan *lua.State, wn+1)
	for i := 0; i < wn; i++ {
		ch <- createLua(options, logger, out)
	}

	return func(next valse.RequestHandler) valse.RequestHandler {
		return func(ctx *valse.Context) error {

			req := createRequest(ctx)
			res := createResponse(ctx)

			L := <-ch
			defer func() {
				go func() {
					ch <- L
				}()
			}()

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
