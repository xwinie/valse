


function fn(req, res)
   
    res.write("Hello, World ")
    return true

end

router:get("/test", function (req, res)
    res.write("hello, world")
end)

router:use(fn)