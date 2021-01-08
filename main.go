package main

import (
	"CryptGuard_Back-end/controllers"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
)

func main() {
	app := iris.New()
	mvc.Configure(app.Party("/user"), setup)
	_ = app.Listen(":8080")
}

func setup(app *mvc.Application) {
	app.Handle(new(controllers.UserController))
}
