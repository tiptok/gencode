package main

import (
	"github.com/tiptok/gencode/cmd"
	_ "github.com/tiptok/gencode/constant"
)

func main() {
	cmd.Init(
		cmd.Name("gencode"),
		cmd.Version("0.0.1"),
		cmd.Description("A tool to gen project"),
	)
}
