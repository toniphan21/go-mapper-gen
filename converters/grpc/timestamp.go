package grpc

import (
	"fmt"
	"go/types"

	"github.com/dave/jennifer/jen"
	gen "github.com/toniphan21/go-mapper-gen"
)

type timestampConverter struct{}

func (c *timestampConverter) Init(parser gen.Parser, config gen.Config) {
	// no-op
}

func (c *timestampConverter) Info() gen.ConverterInfo {
	return gen.ConverterInfo{
		Name:                 "built-in lib grpc/timestampConverter",
		ShortForm:            "timestamppb.Timestamp <-> time.Time",
		ShortFormDescription: "protobuf timestamppb.Timestamp to standard library time.Time",
	}
}

func (c *timestampConverter) CanConvert(ctx gen.ConverterContext, targetType, sourceType types.Type) bool {
	fmt.Println("source", sourceType)
	fmt.Println("target", targetType)
	if c.isTime(sourceType) && c.isTimestamp(targetType) {
		fmt.Println("Time -> Timestamp")
		return true
	}

	if c.isTimestamp(sourceType) && c.isTime(targetType) {
		fmt.Println("Timestamp -> Time")
		return true
	}
	return false
}

func (c *timestampConverter) ConvertField(ctx gen.ConverterContext, target, source gen.Symbol, opts gen.ConverterOption) jen.Code {

	fmt.Println(source.Type)
	fmt.Println(target.Type)
	if c.isTime(source.Type) && c.isTimestamp(target.Type) {

	}

	if c.isTimestamp(source.Type) && c.isTime(target.Type) {
		return source.Expr().Dot("AsTime").Params()
	}
	/**
	func timeToTimestamp(t time.Time) timestamppb.Timestamp {
		return timestamppb.Timestamp{Seconds: int64(t.Unix()), Nanos: int32(t.Nanosecond())}
	}
	*/
	return nil
}

func (c *timestampConverter) isTimestamp(t types.Type) bool {
	const pkgPath = "google.golang.org/protobuf/types/known/timestamppb"
	const typeName = "Timestamp"

	return gen.TypeUtil.MatchNamedType(t, pkgPath, typeName)
}

func (c *timestampConverter) isTime(t types.Type) bool {
	const pkgPath = "time"
	const typeName = "Time"

	return gen.TypeUtil.MatchNamedType(t, pkgPath, typeName)
}

var _ gen.Converter = (*timestampConverter)(nil)
