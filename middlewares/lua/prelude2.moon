

append_all = (buffer, ...) ->
  for i=1,select("#", ...)
    table.insert(buffer, (select(i, ...)))

void_tags = {
  img: true,
  br: true,
  input: true,
}


build_tag = (tag_name, opts) ->
  buffer = { "<", tag_name }
  if type(opts) == "table"
    for k,v in pairs(opts)
      if type(k) ~= "number"
        append_all(buffer, " ", k, '="', v, '"')

  if void_tags[tag_name]
    append_all(buffer, " />")
  else
    append_all(buffer, ">")
    if type(opts) == "table"
      append_all(buffer, unpack(opts))
    else
      append_all(buffer, opts)

    append_all(buffer, "</", tag_name, ">")

  table.concat(buffer)

render_html = (fn) ->
  setfenv fn, setmetatable {},
    __index: (self, tag_name) -> (opts) -> build_tag tag_name, opts

  fn!

json_response = (res) ->
  (data) ->
    out = valse.json.encode data
    res.header.set 'Content-Type', 'application/json'
    res.header.set 'Content-Length', #out
    res.write out

html_response = (res) ->
  (data) ->
    if type data == 'function'
      data = render_html data

    res.header.set 'Content-Type', 'text/html'
    res.header.set 'Content-Length', #out
    res.write out

text_reponse = (res) ->
  (data) ->

    res.header.set 'Content-Type', 'text/plain'
    res.header.set 'Content-Length', #out
    res.write out

response = (res) ->
  setmetatable res,
    __index: (res, key) ->
      switch key
        when "json" then json_response res
        when "html" then html_response res
        else res[key]


class Router
  _routes: {}
  _id: 0
  route: (method, path, fn) =>
    @_id += 1
    @_routes[@_id] = fn
  get: (path, fn) =>
    @route("GET", path, fn)

  
  trigger: (id, req, res) =>
    @_routes[id](req, response(res))
  

export router = Router!

router\get "/test", () ->
  print("hello")

print router._routes