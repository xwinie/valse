package valse

import (
	"bytes"
	"encoding/json"
	"regexp"

	"github.com/kildevaeld/strong"
	"github.com/valyala/fasthttp"
)

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

type Link struct {
	Last    int
	First   int
	Current int
	//Next    int
	//Prev int
	Path []byte
}

var reg = regexp.MustCompile("https?:.*")

const loverheader = 7

func writelink(rel string, url *fasthttp.URI) []byte {
	//l := len(rel) + loverheader + len(url.FullURI())
	//out := make([]byte, l)
	buf := bytes.NewBuffer(nil)
	buf.WriteString("<")
	buf.Write(url.FullURI())
	buf.WriteString(`>; rel="` + rel + `"`)

	return buf.Bytes()
}

func (c *Context) SetLinkHeader(l Link) *Context {
	path := c.Path()
	if l.Path != nil {
		path = l.Path
	}
	url := fasthttp.AcquireURI()
	defer fasthttp.ReleaseURI(url)
	if !reg.Match(path) {
		c.URI().CopyTo(url)
		url.SetPathBytes(path)
	} else {
		url.UpdateBytes(path)
	}

	var links [][]byte
	var page = []byte("page")
	args := url.QueryArgs()

	args.SetUintBytes(page, l.First)
	links = append(links, writelink("first", url))

	args.SetUintBytes(page, l.Current)
	links = append(links, writelink("current", url))

	if l.Last > l.Current {
		args.SetUintBytes(page, l.Current+1)
		links = append(links, writelink("next", url))
	}
	if l.Current > l.First {
		args.SetUintBytes(page, l.Current-1)
		links = append(links, writelink("prev", url))
	}
	args.SetUintBytes(page, l.Last)
	links = append(links, writelink("last", url))
	c.Response.Header.SetBytesV("Link", bytes.Join(links, []byte(", ")))

	return c
}
