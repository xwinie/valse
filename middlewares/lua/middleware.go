package lua

/*
func (l *LuaValse) Middleware() valse.MiddlewareHandler {

	return func(next valse.RequestHandler) valse.RequestHandler {
		return func(ctx *valse.Context) error {

			req := createRequest(ctx)
			res := createResponse(ctx)

			L := <-l.ch
			defer func() {
				go func() {
					l.ch <- L
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
}*/
