package grpc

import gen "github.com/toniphan21/go-mapper-gen"

func RegisterConverters() {
	gen.RegisterConverter(Converters.Timestamp)
}

type converters struct {
	Timestamp gen.Converter
}

var Converters = converters{
	Timestamp: &timestampConverter{},
}
