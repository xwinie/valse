-- class.lua
-- Compatible with Lua 5.1 (not 5.0).
function class(base, init)
   local c = {}    -- a new class instance
   if not init and type(base) == 'function' then
      init = base
      base = nil
   elseif type(base) == 'table' then
    -- our new class is a shallow copy of the base class!
      for i,v in pairs(base) do
         c[i] = v
      end
      c._base = base
   end
   -- the class will be the metatable for all its objects,
   -- and they will look up their methods in it.
   c.__index = c

   -- expose a constructor which can be called by <classname>(<args>)
   local mt = {}
   mt.__call = function(class_tbl, ...)
   local obj = {}
   setmetatable(obj,c)
   if init then
      init(obj,...)
   else 
      -- make sure that any stuff from the base class is initialized!
      if base and base.init then
      base.init(obj, ...)
      end
   end
   return obj
   end
   c.init = init
   c.is_a = function(self, klass)
      local m = getmetatable(self)
      while m do 
         if m == klass then return true end
         m = m._base
      end
      return false
   end
   setmetatable(c, mt)
   return c
end

--[[
local a = {}
local routes = {
	post = {},
	get = {}
}
function register(middleware)
	local l = table.getn(a)
	a[l] = middleware
end

function runMiddlewares(req, res)
	for _, k in pairs(a)  do
		local ret = k(req, res)
		if not ret then
			return false
		end
	end
	return true
end
--]]

Router = class(function(a) 
	a.id = 0
	a.routes = {}
end)


function Router:route(method, path, fn)
	self.id = self.id + 1
	self.routes[self.id] = fn
	__create_route(method, path, self.id)
end

function Router:get(path, fn)
	self:route("GET", path, fn)
end

function Router:post(path, fn)
	self:route("POST", path, fn)
end

function Router:put(path, fn)
	self:route("PUT", path, fn)
end

function Router:delete(path, fn)
	self:route("DELETE", path, fn)
end

function Router:head(path, fn) 
	self:route("HEAD", path, fn)
end

function Router:use(fn)
	self.id = self.id + 1
	self.routes[self.id] = fn
	__create_middleware(self.id)
end

function Router:trigger(id, req, res)
	return self.routes[id](req, res)
end


router = Router()