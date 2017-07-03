package main

import (
	"github.com/kildevaeld/valse"
	"github.com/kildevaeld/valse/middlewares/lua"
)

func main() {
	//logrus.SetLevel(logrus.DebugLevel)
	server := valse.New()

	l := lua.New(server, lua.LuaOptions{
		Path:      ".",
		WorkQueue: 20,
	})

	l.Open()
	defer l.Close()

	server.Listen(":3000")

}
