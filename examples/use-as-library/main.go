package main

import (
	"fmt"

	gomappergen "github.com/toniphan21/go-mapper-gen"
)

func main() {
	fmt.Println("Use go-mapper-gen as library")

	gomappergen.RegisterAllBuiltinConverters()

	var configs []gomappergen.Config
	_ = gomappergen.Generate(nil, nil, configs)
}
