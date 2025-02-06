package main

import (
	"ttpos-server-go/migration/cmd"

	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	cmd.Execute()
}
