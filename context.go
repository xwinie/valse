package valse

import (
	"github.com/json-iterator/go"
	"github.com/kildevaeld/strong"
	"github.com/valyala/fasthttp"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

// Context represents the context of the HTTP request.
type Context struct {
	noCopy
	*fasthttp.RequestCtx
	log Logger
	s   *Server
}

func (c *Context) reset() *Context {
	c.RequestCtx = nil
	return c
}

// Log log function
func (c *Context) Log() Logger {
	return c.log
}

// Status sets the response status code.
func (c *Context) Status(status int) *Context {
	c.SetStatusCode(status)
	return c
}

// JSON marshal the Object to JSON and writes it via the ResponseWriter.
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

// Text marshal the string to JSON and writes it via the ResponseWriter.
func (c *Context) Text(v string) error {
	c.SetStatusCode(strong.StatusOK)
	c.Response.Header.Set(strong.HeaderContentType, strong.MIMETextPlainCharsetUTF8)
	c.Response.SetBodyString(v)
	return nil
}

// PathParameter accesses the Path parameter value by its name
func (c *Context) PathParameter(name string) string {
	return c.UserValue(name).(string)
}

// QueryParameter returns the (first) Query parameter value by its name
func (c *Context) QueryParameter(name string) []byte {
	return c.FormValue(name)
}

// BodyParameter parses the body of the request (once for typically a POST or a PUT) and returns the value of the given name or an error.
func (c *Context) BodyParameter(name string) []byte {
	return c.FormValue(name)
}

// GetJSONObject call json.Unmarshal by sending the reference of the given object.
func (c *Context) GetJSONObject(object interface{}) error {
	return json.Unmarshal(c.PostBody(), &object)
}

// GetBody read post put and delete request body return body []byte
func (c *Context) GetBody() []byte {
	return c.PostBody()
}

// HeaderParameter returns the HTTP Header value of a Header name or empty if missing
func (c *Context) HeaderParameter(name string) []byte {
	return c.Request.Header.Peek(name)
}
