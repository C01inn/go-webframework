# Create Web Apis in Go Quickly


### Hello World App

```
package main

import "github.com/C01inn/go-webframework"

func main () {
    app := Server()

    app.get("/", func(req req) urlResp {
    	return sendStr("Hello World!")
    })
    
    app.listen(8080)
}
```

### Route with get request
```
app.get("/", func(req, req) urlResp {
    return sendStr("Hello World")  
})
```

### Route with post request
```
app.post("/", func(req, req) urlResp {
    return sendStr("Hello World")  
})
```

## Response Data Types
##
### Send Text
```
app.post("/", func(req, req) urlResp {
    return sendStr("Hello World")  
})
```
### Render an Html File

```
app.get("/", func(req, req) urlResp {
    return renderHtml(filepath, nil)
})
```

### Html Templating
#### Create a struct of your data and pass it to the Html Template
#### If you wish to send no data simply pass a nil value
```
app.get("/", func(req req) urlResp {
    // create a struct of data to pass to html
    type mydata struct {
        Title   string
        Name    string
    }
    data := mydata{
        Title: "My title",
        Name: "My name",
    }

    return renderHtml(filepath, data)
})
```

### Send Json Data
###### Use the sendJson() function to send json data. The function takes in a map, slice, array, or json string.
##
```
app.get("/", func(req, req) urlResp {
    return sendJson(data)
})
```

### Send a static file
```
app.get("/", func(req, req) urlResp {
    return sendFile(filepath)
})
```
## Send a downloadable file
```
app.get("/", func(req, req) urlResp {
    return downloadFile(filepath, filename)
})
```
***






