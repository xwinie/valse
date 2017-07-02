
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

	local r
	if req.method == "GET" then
		r = routes.get
	end
    if req.method == "POST" then
        r = routes.post
    end


	if r then
		for path, route in pairs(r) do
			if path == req.path then
				if not route(req, res) then
					return false
				end
			end
		end
	end

	for _, k in pairs(a)  do
		local ret = k(req, res)
		if not ret then
			return false
		end
	end
	return true
end

router = {}

function router.get(path, fn)
	routes.get[path] = fn
end

function router.post(path, fn)
    routes.post[path] = fn
end