package main

import (
	"github.com/cristiano-pacheco/go-online-auction/cmd"
	"github.com/cristiano-pacheco/go-online-auction/internal/shared/modules/config"
)

func main() {
	config.Init()
	cmd.Execute()
}
