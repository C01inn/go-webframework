package main

import (
	"swift"
)

func main() {

	app := swift.Server()

	app.Get("/", func(req swift.Req) swift.UrlResp {
		return swift.SendStr("Hello World!")
	})

	app.Listen(6060)
}
