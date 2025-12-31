// Code generated from Pkl module `gomappergen.mapper`. DO NOT EDIT.
package mapper

import (
	"context"

	"github.com/apple/pkl-go/pkl"
)

type Mapper struct {
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a Mapper
func LoadFromPath(ctx context.Context, path string) (ret Mapper, err error) {
	evaluator, err := pkl.NewEvaluator(ctx, pkl.PreconfiguredOptions)
	if err != nil {
		return ret, err
	}
	defer func() {
		cerr := evaluator.Close()
		if err == nil {
			err = cerr
		}
	}()
	ret, err = Load(ctx, evaluator, pkl.FileSource(path))
	return ret, err
}

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a Mapper
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (Mapper, error) {
	var ret Mapper
	err := evaluator.EvaluateModule(ctx, source, &ret)
	return ret, err
}
