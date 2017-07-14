local append_all
append_all = function(buffer, ...)
  for i = 1, select("#", ...) do
    table.insert(buffer, (select(i, ...)))
  end
end
local void_tags = {
  img = true,
  br = true,
  input = true
}
local build_tag
build_tag = function(tag_name, opts)
  local buffer = {
    "<",
    tag_name
  }
  if type(opts) == "table" then
    for k, v in pairs(opts) do
      if type(k) ~= "number" then
        append_all(buffer, " ", k, '="', v, '"')
      end
    end
  end
  if void_tags[tag_name] then
    append_all(buffer, " />")
  else
    append_all(buffer, ">")
    if type(opts) == "table" then
      append_all(buffer, unpack(opts))
    else
      append_all(buffer, opts)
    end
    append_all(buffer, "</", tag_name, ">")
  end
  return table.concat(buffer)
end
local render_html
render_html = function(fn)
  setfenv(fn, setmetatable({ }, {
    __index = function(self, tag_name)
      return function(opts)
        return build_tag(tag_name, opts)
      end
    end
  }))
  return fn()
end
local json_response
json_response = function(res)
  return function(data)
    local out = valse.json.encode(data)
    res.header.set('Content-Type', 'application/json')
    res.header.set('Content-Length', #out)
    return res.write(out)
  end
end
local html_response
html_response = function(res)
  return function(data)
    if type(data == 'function') then
      data = render_html(data)
    end
    res.header.set('Content-Type', 'text/html')
    res.header.set('Content-Length', #out)
    return res.write(out)
  end
end
local text_reponse
text_reponse = function(res)
  return function(data)
    res.header.set('Content-Type', 'text/plain')
    res.header.set('Content-Length', #out)
    return res.write(out)
  end
end
local response
response = function(res)
  return setmetatable(res, {
    __index = function(res, key)
      local _exp_0 = key
      if "json" == _exp_0 then
        return json_response(res)
      elseif "html" == _exp_0 then
        return html_response(res)
      else
        return res[key]
      end
    end
  })
end
local Router
do
  local _class_0
  local _base_0 = {
    _routes = { },
    _id = 0,
    route = function(self, method, path, fn)
      self._id = self._id + 1
      self._routes[self._id] = fn
    end,
    get = function(self, path, fn)
      return self:route("GET", path, fn)
    end,
    trigger = function(self, id, req, res)
      return self._routes[id](req, response(res))
    end
  }
  _base_0.__index = _base_0
  _class_0 = setmetatable({
    __init = function() end,
    __base = _base_0,
    __name = "Router"
  }, {
    __index = _base_0,
    __call = function(cls, ...)
      local _self_0 = setmetatable({}, _base_0)
      cls.__init(_self_0, ...)
      return _self_0
    end
  })
  _base_0.__class = _class_0
  Router = _class_0
end
function print_r ( t )  
    local print_r_cache={}
    local function sub_print_r(t,indent)
        if (print_r_cache[tostring(t)]) then
            print(indent.."*"..tostring(t))
        else
            print_r_cache[tostring(t)]=true
            if (type(t)=="table") then
                for pos,val in pairs(t) do
                    if (type(val)=="table") then
                        print(indent.."["..pos.."] => "..tostring(t).." {")
                        sub_print_r(val,indent..string.rep(" ",string.len(pos)+8))
                        print(indent..string.rep(" ",string.len(pos)+6).."}")
                    elseif (type(val)=="string") then
                        print(indent.."["..pos..'] => "'..val..'"')
                    else
                        print(indent.."["..pos.."] => "..tostring(val))
                    end
                end
            else
                print(indent..tostring(t))
            end
        end
    end
    if (type(t)=="table") then
        print(tostring(t).." {")
        sub_print_r(t,"  ")
        print("}")
    else
        sub_print_r(t,"  ")
    end
    print()
end


router = Router()
router:get("/test", function()
  return print("hello")
end)

router:get("/test", function()
  return print("hello")
end)
return print_r(router._routes)
