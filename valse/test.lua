
local json = require 'json'

function fn(req, res)
   
    res.write("Hello, World ")
    return true

end

router:get("/test", function (req, res)
    local str, err = valse.json.encode({
        rapper = "er lige med"
    })
    res.header.set('Content-Type', 'application/json')
    res.write(str)
end)

--router:use(fn)