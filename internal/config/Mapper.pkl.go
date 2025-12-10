// Code generated from Pkl module `gomappergen.Config`. DO NOT EDIT.
package config

type Mapper interface {
	BaseMapper

	GetPriorities() *map[int]BaseMapper
}

var _ Mapper = MapperImpl{}

type MapperImpl struct {
	BaseMapperImpl

	Priorities *map[int]BaseMapper `pkl:"priorities"`
}

func (rcv MapperImpl) GetPriorities() *map[int]BaseMapper {
	return rcv.Priorities
}
