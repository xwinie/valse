package lua

import (
	"github.com/kildevaeld/valse"
	"github.com/stevedonovan/luar"
)

func execute(ctx *valse.Context, ch chan *VM, id int, middleware bool) bool {
	req := createRequest(ctx)
	res := createResponse(ctx)

	vm := <-ch
	defer func() {
		ch <- vm
	}()

	/*logrus.WithFields(logrus.Fields{
		"path": string(ctx.Path()),
		"vm":   vm.id,
	}).Debugf("execute lua script")*/

	state := vm.state
	state.GetGlobal("router")
	state.GetGlobal("Router")
	state.GetField(-1, "trigger")
	state.PushValue(-3)
	luar.GoToLua(state, id)
	luar.GoToLua(state, req)
	luar.GoToLua(state, res)
	state.Call(4, 1)

	return vm.state.ToBoolean(0)
}

func route(id int, ch chan *VM) valse.RequestHandler {

	return func(ctx *valse.Context) error {

		execute(ctx, ch, id, false)

		return nil
	}
}

func middleware(id int, ch chan *VM) valse.MiddlewareHandler {
	return func(next valse.RequestHandler) valse.RequestHandler {
		return func(ctx *valse.Context) error {

			if execute(ctx, ch, id, true) {
				return next(ctx)
			}
			return nil
		}
	}
}
