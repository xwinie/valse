
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

router:get("/", function (req, res) 

    res.header.set("Content-Type", "text/html")
    print("here", render_html)
    res.write(render_html(function()
        return div {
            h1 "Hello, World"
        }
    end))

end)

--router:use(fn)