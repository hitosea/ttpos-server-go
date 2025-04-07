package main

import (
	"runtime"
	"ttpos-server-go/command"
)

// @title ttpos-server-go API
// @version 1.0
// @description 点餐系统服务端接口文档

// @securityDefinitions.apikey JwtToken
// @in header
// @name Authorization

// @host 127.0.0.1:8080
// @BasePath /
func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	command.Execute()
}
