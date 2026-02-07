package main

import (
	"github.com/herewei/ohmymem-core/cmd"
	_ "github.com/herewei/ohmymem-core/cmd/init"
	_ "github.com/herewei/ohmymem-core/cmd/mcp"
)

func main() {
	cmd.Execute()
}
