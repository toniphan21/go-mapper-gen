package main

import (
	"github.com/alexflint/go-arg"
	"github.com/toniphan21/go-mapper-gen/internal/cli"
)

func main() {
	var args cli.Args
	arg.MustParse(&args)

	cli.Run(args)
}
