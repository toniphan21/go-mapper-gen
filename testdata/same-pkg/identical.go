package domain

import "time"

type IdenticalTarget struct {
	Bool    bool
	Int     int
	Int8    int8
	Int16   int16
	Int32   int32
	Int64   int64
	Float32 float32
	Float64 float64
	Complex complex64
	String  string
	Time    time.Time

	PointerBool    *bool
	PointerInt     *int
	PointerInt8    *int8
	PointerInt16   *int16
	PointerInt32   *int32
	PointerInt64   *int64
	PointerFloat32 *float32
	PointerFloat64 *float64
	PointerComplex *complex64
	PointerString  *string
	PointerTime    *time.Time
}

type IdenticalSource struct {
	Bool    bool
	Int     int
	Int8    int8
	Int16   int16
	Int32   int32
	Int64   int64
	Float32 float32
	Float64 float64
	Complex complex64
	String  string
	Time    time.Time

	PointerBool    *bool
	PointerInt     *int
	PointerInt8    *int8
	PointerInt16   *int16
	PointerInt32   *int32
	PointerInt64   *int64
	PointerFloat32 *float32
	PointerFloat64 *float64
	PointerComplex *complex64
	PointerString  *string
	PointerTime    *time.Time
}
