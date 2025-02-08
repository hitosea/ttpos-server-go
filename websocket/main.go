package main

import (
	"runtime"
	"websocket/command"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	command.Execute()
}
