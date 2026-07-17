package main

import (
	"auction/cmd"
	"auction/internal/shared/modules/config"
)

func main() {
	config.Init()
	cmd.Execute()
}
