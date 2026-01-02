package gomappergen

import (
	"fmt"
	"go/types"
)

var LookUpTotalHits uint64

var enableLookUpCache = false

func EnableLookUpCache() {
	enableLookUpCache = true
}

type Descriptor struct {
	structInfo      *StructInfo
	structFieldInfo *StructFieldInfo
}

func (d *Descriptor) StructType() types.Type {
	if d.structInfo == nil {
		return nil
	}
	return d.structInfo.Type
}

func (d *Descriptor) StructFields() map[string]StructFieldInfo {
	if d.structInfo == nil {
		return nil
	}
	return d.structInfo.Fields
}

func (d *Descriptor) FieldName() string {
	if d.structFieldInfo == nil {
		return ""
	}
	return d.structFieldInfo.Name
}

func (d *Descriptor) FieldType() types.Type {
	if d.structFieldInfo == nil {
		return nil
	}
	return d.structFieldInfo.Type
}

func (d *Descriptor) FieldIndex() int {
	if d.structFieldInfo == nil {
		return -1
	}
	return d.structFieldInfo.Index
}

type lookUpCache struct {
	converter Converter
	target    types.Type
	source    types.Type
}

type LookupContext interface {
	// LookUp searches the global converter registry for a converter that
	// can convert a value of sourceType to targetType, excluding the provided
	// currentConverter (if non-nil).
	//
	// This helper is intended for converter implementations that need to
	// delegate or reuse existing conversion rules. A common use-case is a
	// SliceConverter that converts []T -> []V by looking up a converter for
	// T -> V and then generating per-element conversion code.
	//
	// Selection rules (implementation contract):
	//  1. The registry is scanned for converters c where c.CanConvert(targetType, sourceType)
	//     returns true.
	//  2. The currentConverter parameter is excluded from consideration to avoid
	//     trivial self-selection (if currentConverter == nil, no exclusion occurs).
	//  3. From the remaining candidates, the converter with the highest priority
	//     (your package's ordering rule: lower numeric value = higher priority)
	//     is chosen. If multiple converters share the same priority, the selection
	//     must be deterministic (for example: registration order or stable sorting).
	//
	// Return value:
	//   - (Converter, true) if a matching converter was found.
	//   - (nil, false) if no converter in the registry can perform the conversion.
	LookUp(current Converter, targetType, sourceType types.Type) (Converter, error)

	TargetDescriptor() *Descriptor

	SourceDescriptor() *Descriptor
}

type lookupContext struct {
	converters []*registeredConverter
	target     Descriptor
	source     Descriptor
}

func newLookupContext(target Descriptor, source Descriptor) *lookupContext {
	return &lookupContext{
		target: target,
		source: source,
	}
}

func emptyLookupContext() *lookupContext {
	return &lookupContext{}
}

var lkCache []lookUpCache

func (l *lookupContext) LookUp(current Converter, targetType, sourceType types.Type) (Converter, error) {
	if current == nil {
		return nil, fmt.Errorf("invalid: current converter is nil")
	}

	var reachable []*registeredConverter
	var available []*registeredConverter
	if l.converters == nil {
		available = globalConverters
	} else {
		available = l.converters
	}

	for _, reg := range available {
		if current == reg.converter {
			continue
		}

		if normalizeConverterType(current) == reg.typ {
			continue
		}

		reachable = append(reachable, reg)
	}

	ctx := &lookupContext{
		converters: reachable,
		target:     l.target,
		source:     l.source,
	}
	var nextContext []string
	for _, converter := range ctx.converters {
		nextContext = append(nextContext, converter.typ.Name())
	}

	if enableLookUpCache {
		for _, c := range lkCache {
			if TypeUtil.IsIdentical(targetType, c.target) && TypeUtil.IsIdentical(sourceType, c.source) {
				for _, reg := range reachable {
					if c.converter == reg.converter {
						return c.converter, nil
					}
				}
			}
		}
	}

	for _, reg := range reachable {
		LookUpTotalHits++
		if reg.converter.CanConvert(ctx, targetType, sourceType) {
			if enableLookUpCache {
				lkCache = append(lkCache, lookUpCache{
					converter: reg.converter,
					target:    targetType,
					source:    sourceType,
				})
			}
			return reg.converter, nil
		}
	}
	return nil, fmt.Errorf("unable to find matching converter for target %s, source %s", targetType.String(), sourceType.String())
}

func (l *lookupContext) TargetDescriptor() *Descriptor {
	return &l.target
}

func (l *lookupContext) SourceDescriptor() *Descriptor {
	return &l.source
}

var _ LookupContext = (*lookupContext)(nil)

func findConverter(target, source Descriptor) (Converter, bool) {
	for _, reg := range globalConverters {
		LookUpTotalHits++
		if reg.converter.CanConvert(newLookupContext(target, source), target.structFieldInfo.Type, source.structFieldInfo.Type) {
			if enableLookUpCache {
				lkCache = append(lkCache, lookUpCache{
					converter: reg.converter,
					target:    target.structFieldInfo.Type,
					source:    source.structFieldInfo.Type,
				})
			}
			return reg.converter, true
		}
	}
	return nil, false
}
