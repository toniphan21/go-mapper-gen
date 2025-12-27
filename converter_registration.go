package gomappergen

import (
	"fmt"
	"log/slog"
	"maps"
	"reflect"
	"slices"
	"sort"
	"strconv"

	"github.com/IGLOU-EU/go-wildcard"
	"github.com/toniphan21/go-mapper-gen/internal/util"
)

// RegisterConverter adds a new Converter to the global converter registry.
//
// Converters are selected by the mapper generator based on:
//  1. Whether they report true from CanConvert(...)
//  2. Their assigned priority (use converter { priority = List(...) } config)
//
// The priority determines via converter { priority } configuration. The
// ordering between converters that can handle the same source and target types.
//
// Registering a converter does not trigger any generation work. It merely
// appends the converter to the internal registry so that the code generator
// can discover it later.
//
// Converters should be stateless and safe for repeated reuse. The registry
// is typically read during generator initialization or when resolving the
// appropriate converter for a given assignment.
//
// Example:
//
//	RegisterConverter(&StringToIntConverter{})
//
// Passing the same converter instance multiple times is allowed but generally
// discouraged unless intentional.
func RegisterConverter(converter Converter) {
	globalConverters = append(globalConverters, &registeredConverter{
		converter:     converter,
		typ:           normalizeConverterType(converter),
		qualifiedName: qualifiedName(converter),
		priority:      len(globalConverters) + 1,
		builtIn:       false,
	})
}

type registeredConverter struct {
	converter     Converter
	typ           reflect.Type
	priority      int
	builtIn       bool
	qualifiedName string
}

var globalConverters []*registeredConverter

func normalizeConverterType(v Converter) reflect.Type {
	if v == nil {
		return nil
	}
	t := reflect.TypeOf(v)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

func qualifiedName(v Converter) string {
	t := reflect.TypeOf(v)
	if t == nil {
		return "<nil>"
	}

	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	pkg := t.PkgPath()
	name := t.Name()

	if pkg == "" {
		return name
	}
	return pkg + "." + name
}

func matchConverterPriority(config []string, qualifiedName string) (bool, int) {
	var matched = make(map[int]int) // index -> pattern-length
	for i, pattern := range config {
		if pattern == qualifiedName {
			matched[i] = len(pattern)
			continue
		}

		if wildcard.Match(pattern, qualifiedName) {
			matched[i] = len(pattern)
		}
	}

	if len(matched) == 0 {
		return false, -1
	}

	// longest pattern always win
	var result, l int
	for idx, length := range matched {
		if length > l {
			result = idx
			l = length
		}
	}
	return true, result
}

func getOrderConverterQualifiedNamesByPrioritiesConfig(regs []*registeredConverter, config []string) []string {
	var m = make(map[int][]string)

	for idx, reg := range regs {
		matched, priority := matchConverterPriority(config, reg.qualifiedName)
		if matched {
			m[priority] = append(m[priority], reg.qualifiedName)
		} else {
			m[idx] = append(m[idx], reg.qualifiedName)
		}
	}

	var result []string
	keys := slices.Collect(maps.Keys(m))
	sort.Ints(keys)
	for _, k := range keys {
		v := m[k]
		sort.Strings(v)
		for _, s := range v {
			result = append(result, s)
		}
	}
	return result
}

func prioritizeRegisteredConverters(parsedConfig Config) {
	var prioritied = getOrderConverterQualifiedNamesByPrioritiesConfig(globalConverters, parsedConfig.ConverterPriorities)
	for i, v := range prioritied {
		for _, reg := range globalConverters {
			if reg.qualifiedName == v {
				reg.priority = i
			}
		}
	}

	slices.SortFunc(globalConverters, func(a, b *registeredConverter) int {
		return a.priority - b.priority
	})
}

func initRegisteredConverters(parser Parser, parsedConfig Config) {
	prioritizeRegisteredConverters(parsedConfig)
	for _, v := range globalConverters {
		v.converter.Init(parser, parsedConfig)
	}
}

func registerBuiltInConverter(converter Converter, priority int) {
	globalConverters = append(globalConverters, &registeredConverter{
		converter:     converter,
		typ:           normalizeConverterType(converter),
		qualifiedName: qualifiedName(converter),
		priority:      priority,
		builtIn:       true,
	})

	slices.SortFunc(globalConverters, func(a, b *registeredConverter) int {
		return a.priority - b.priority
	})
}

func PrintRegisteredConverters(logger *slog.Logger) {
	shortFormBuffer := 0
	for _, c := range globalConverters {
		info := c.converter.Info()
		l := len(info.ShortForm)
		if l > shortFormBuffer {
			shortFormBuffer = l
		}
	}

	for _, v := range globalConverters {
		builtin := ""
		if v.builtIn {
			builtin = util.ColorCyan("[built-in]")
		}

		info := v.converter.Info()
		line := fmt.Sprintf("%-*s %v", shortFormBuffer+1, info.ShortForm, info.ShortFormDescription)

		logger.Info(
			fmt.Sprintf("%s %10s %s",
				util.ColorBlue(fmt.Sprintf("%5s", strconv.Itoa(v.priority))),
				builtin,
				line,
			),
		)
	}
}

func ClearAllRegisteredConverters() {
	globalConverters = []*registeredConverter{}
}

func RegisterAllBuiltinConverters() {
	cf := BuiltInConverterConfig{}
	cf.EnableAll()

	RegisterBuiltinConverters(cf)
}

func RegisterBuiltinConverters(config BuiltInConverterConfig) {
	priority := 0
	if config.UseIdentical {
		registerBuiltInConverter(BuiltinConverters.IdenticalType, 0)
	}

	if config.UseSlice {
		registerBuiltInConverter(BuiltinConverters.Slice, priority)
		priority++
	}

	if config.UseTypeToPointer {
		registerBuiltInConverter(BuiltinConverters.TypeToPointer, priority)
		priority++
	}

	if config.UsePointerToType {
		registerBuiltInConverter(BuiltinConverters.PointerToType, priority)
		priority++
	}

	if config.UseNumeric {
		registerBuiltInConverter(BuiltinConverters.Numeric, priority)
		priority++
	}

	if config.UseFunctions {
		registerBuiltInConverter(BuiltinConverters.Functions, priority)
		priority++
	}
}

type builtinConverters struct {
	IdenticalType Converter
	Slice         Converter
	TypeToPointer Converter
	PointerToType Converter
	Numeric       Converter
	Functions     Converter
}

var BuiltinConverters = builtinConverters{
	IdenticalType: &identicalTypeConverter{},
	Slice:         &sliceConverter{},
	TypeToPointer: &typeToPointerConverter{},
	PointerToType: &pointerToTypeConverter{},
	Numeric:       &numericConverter{},
	Functions:     &functionsConverter{},
}
