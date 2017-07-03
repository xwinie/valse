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

type VM struct {
	state *lua.State
	id    int
}

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
	LuaFactory  func() *lua.State
	WorkQueue   int
}

type File struct {
	Path    string
	Content string
}

type RouterFactory func(method, path string, id int)

func createLua(options LuaOptions, logger logrus.FieldLogger, files []File, factory RouterFactory) *lua.State {
	var L *lua.State

	if options.LuaFactory != nil {
		L = options.LuaFactory()
	} else {
		L = luar.Init()
		L.OpenLibs()
	}

	L.Register("__create_route", func(state *lua.State) int {
		method := state.ToString(1)
		route := state.ToString(2)
		id := state.ToInteger(3)

		factory(method, route, id)
		return 0
	})

	L.Register("__create_middleware", func(state *lua.State) int {
		id := state.ToInteger(1)
		factory("", "", id)
		return 0
	})

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

type LuaValse struct {
	o  LuaOptions
	ch chan *lua.State
	s  *valse.Server
	id int
}

func (l LuaValse) loadFiles() []File {
	logger := logrus.WithField("prefix", "middlewares:lua")

	files, err := getSortedFiles(l.o.Path)
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
	return out
}

func (l *LuaValse) Open() {
	wn := l.o.WorkQueue
	if wn == 0 {
		wn = 5
	}

	logger := logrus.WithField("prefix", "middleware:lua")

	files := l.loadFiles()

	ch := make(chan *VM, wn+1)
	for i := 0; i < wn; i++ {
		lua := createLua(l.o, logger, files, func(method, path string, id int) {
			if id <= l.id {
				return
			}
			l.id = id
			if method == "" {
				l.s.Use(middleware(id, ch))
			} else {
				l.s.Route(method, path, route(id, ch))
			}
		})

		ch <- &VM{lua, i}
	}

}

func (l *LuaValse) Close() {
	for c := range l.ch {
		c.Close()
	}
}
func New(server *valse.Server, o LuaOptions) *LuaValse {
	return &LuaValse{o, nil, server, 0}
}
