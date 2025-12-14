package domain

type iMapperIdentical interface {
	// ToIdenticalTarget converts a IdenticalSource value into a IdenticalTarget value.
	ToIdenticalTarget(in IdenticalSource) IdenticalTarget

	// FromIdenticalTarget converts a IdenticalTarget value into a IdenticalSource
	// value.
	FromIdenticalTarget(in IdenticalTarget) IdenticalSource

	// ToTargetWithPointerNone converts a Source value into a Target value.
	ToTargetWithPointerNone(in Source) Target

	// FromTargetWithPointerNone converts a Target value into a Source value.
	FromTargetWithPointerNone(in Target) Source

	// ToTargetWithPointerTargetOnly converts a Source value into a Target value.
	ToTargetWithPointerTargetOnly(in Source) *Target

	// FromTargetWithPointerTargetOnly converts a Target value into a Source value.
	FromTargetWithPointerTargetOnly(in *Target) Source

	// ToTargetWithPointerSourceOnly converts a Source value into a Target value.
	ToTargetWithPointerSourceOnly(in *Source) Target

	// FromTargetWithPointerSourceOnly converts a Target value into a Source value.
	FromTargetWithPointerSourceOnly(in Target) *Source

	// ToTargetWithPointerBoth converts a Source value into a Target value.
	ToTargetWithPointerBoth(in *Source) *Target

	// FromTargetWithPointerBoth converts a Target value into a Source value.
	FromTargetWithPointerBoth(in *Target) *Source

	// ToTargetWithoutFrom converts a Source value into a Target value.
	ToTargetWithoutFrom(in Source) Target

	// FromTargetWithoutTo converts a Target value into a Source value.
	FromTargetWithoutTo(in Target) Source
}

type iMapperIdenticalImpl struct{}

func (m *iMapperIdenticalImpl) ToIdenticalTarget(in IdenticalSource) IdenticalTarget {
	var out IdenticalTarget

	out.Bool = in.Bool
	out.Complex = in.Complex
	out.Float32 = in.Float32
	out.Float64 = in.Float64
	out.Int = in.Int
	out.Int16 = in.Int16
	out.Int32 = in.Int32
	out.Int64 = in.Int64
	out.Int8 = in.Int8
	out.PointerBool = in.PointerBool
	out.PointerComplex = in.PointerComplex
	out.PointerFloat32 = in.PointerFloat32
	out.PointerFloat64 = in.PointerFloat64
	out.PointerInt = in.PointerInt
	out.PointerInt16 = in.PointerInt16
	out.PointerInt32 = in.PointerInt32
	out.PointerInt64 = in.PointerInt64
	out.PointerInt8 = in.PointerInt8
	out.PointerString = in.PointerString
	out.PointerTime = in.PointerTime
	out.String = in.String
	out.Time = in.Time

	return out
}

func (m *iMapperIdenticalImpl) FromIdenticalTarget(in IdenticalTarget) IdenticalSource {
	var out IdenticalSource

	out.Bool = in.Bool
	out.Complex = in.Complex
	out.Float32 = in.Float32
	out.Float64 = in.Float64
	out.Int = in.Int
	out.Int16 = in.Int16
	out.Int32 = in.Int32
	out.Int64 = in.Int64
	out.Int8 = in.Int8
	out.PointerBool = in.PointerBool
	out.PointerComplex = in.PointerComplex
	out.PointerFloat32 = in.PointerFloat32
	out.PointerFloat64 = in.PointerFloat64
	out.PointerInt = in.PointerInt
	out.PointerInt16 = in.PointerInt16
	out.PointerInt32 = in.PointerInt32
	out.PointerInt64 = in.PointerInt64
	out.PointerInt8 = in.PointerInt8
	out.PointerString = in.PointerString
	out.PointerTime = in.PointerTime
	out.String = in.String
	out.Time = in.Time

	return out
}

func (m *iMapperIdenticalImpl) ToTargetWithPointerNone(in Source) Target {
	var out Target

	out.ID = in.ID

	return out
}

func (m *iMapperIdenticalImpl) FromTargetWithPointerNone(in Target) Source {
	var out Source

	out.ID = in.ID

	return out
}

func (m *iMapperIdenticalImpl) ToTargetWithPointerTargetOnly(in Source) *Target {
	var out Target

	out.ID = in.ID

	return &out
}

func (m *iMapperIdenticalImpl) FromTargetWithPointerTargetOnly(in *Target) Source {
	var out Source

	out.ID = in.ID

	return out
}

func (m *iMapperIdenticalImpl) ToTargetWithPointerSourceOnly(in *Source) Target {
	var out Target

	out.ID = in.ID

	return out
}

func (m *iMapperIdenticalImpl) FromTargetWithPointerSourceOnly(in Target) *Source {
	var out Source

	out.ID = in.ID

	return &out
}

func (m *iMapperIdenticalImpl) ToTargetWithPointerBoth(in *Source) *Target {
	var out Target

	out.ID = in.ID

	return &out
}

func (m *iMapperIdenticalImpl) FromTargetWithPointerBoth(in *Target) *Source {
	var out Source

	out.ID = in.ID

	return &out
}

func (m *iMapperIdenticalImpl) ToTargetWithoutFrom(in Source) Target {
	var out Target

	out.ID = in.ID

	return out
}

func (m *iMapperIdenticalImpl) FromTargetWithoutTo(in Target) Source {
	var out Source

	out.ID = in.ID

	return out
}

var _ iMapperIdentical = (*iMapperIdenticalImpl)(nil)
