package grpc

import gen "github.com/toniphan21/go-mapper-gen"

func RegisterConverters() {
	gen.RegisterConverter(&timestampConverter{}, gen.RegisteredConverterCount())
}
