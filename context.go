package valse

import (
	"encoding/json"

	"github.com/kildevaeld/strong"
	"github.com/valyala/fasthttp"
)

type Context struct {
	noCopy
	*fasthttp.RequestCtx
	log Logger
}

func (c *Context) reset() *Context {
	c.RequestCtx = nil
	return c
}

func (c *Context) Log() Logger {
	return c.log
}

func (c *Context) Status(status int) *Context {
	c.SetStatusCode(status)
	return c
}

func (c *Context) JSON(v interface{}) error {
	c.SetStatusCode(strong.StatusOK)
	c.Response.Header.Set(strong.HeaderContentType, strong.MIMEApplicationJSONCharsetUTF8)
	bs, err := json.Marshal(v)
	if err != nil {
		return err
	}
	c.Response.SetBody(bs)
	return nil
}

func (c *Context) Text(v string) error {
	c.SetStatusCode(strong.StatusOK)
	c.Response.Header.Set(strong.HeaderContentType, strong.MIMETextPlainCharsetUTF8)
	c.Response.SetBodyString(v)
	return nil
}
