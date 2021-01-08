package main

import (
	"./controllers"
	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
)

func main() {
	app := iris.New()
	mvc.Configure(app.Party("/user"), setup)
	_ = app.Listen(":8080")
}

func setup(app *mvc.Application) {
	app.Handle(new(controllers.UserController))
}
