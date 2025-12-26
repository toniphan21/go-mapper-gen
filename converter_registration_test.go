package gomappergen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_matchConverterPriority(t *testing.T) {
	cases := []struct {
		name          string
		config        []string
		qualifiedName string
		expectedMatch bool
		expectedIndex int
	}{
		{
			name:          "empty config resulted no match",
			config:        []string{},
			qualifiedName: "anything",
			expectedMatch: false,
			expectedIndex: -1,
		},

		{
			name:          "not matched",
			config:        []string{"github.com/toniphan21/go-mapper-gen.identicalTypeConverter"},
			qualifiedName: "anything",
			expectedMatch: false,
			expectedIndex: -1,
		},

		{
			name: "matched by name",
			config: []string{
				"github.com/toniphan21/go-mapper-gen.identicalTypeConverter",
				"github.com/toniphan21/go-mapper-gen.functionsConverter",
				"github.com/toniphan21/go-mapper-gen.numericConverter",
			},
			qualifiedName: "github.com/toniphan21/go-mapper-gen.functionsConverter",
			expectedMatch: true,
			expectedIndex: 1,
		},

		{
			name: "matched by pattern",
			config: []string{
				"github.com/toniphan21/go-mapper-gen.identicalTypeConverter",
				"github.com/toniphan21/go-mapper-gen.functionsConverter",
				"github.com/other/lib.*",
				"github.com/toniphan21/go-mapper-gen.numericConverter",
			},
			qualifiedName: "github.com/other/lib.customConverter",
			expectedMatch: true,
			expectedIndex: 2,
		},

		{
			name: "multiple matches, longest win",
			config: []string{
				"github.com/other/lib.*",
				"*",
				"github.com/other/*",
				"github.com/other/lib.customConverter",
				"github.com/other/lib.custom*",
			},
			qualifiedName: "github.com/other/lib.customConverter",
			expectedMatch: true,
			expectedIndex: 3,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			match, index := matchConverterPriority(tc.config, tc.qualifiedName)

			assert.Equal(t, tc.expectedMatch, match)
			assert.Equal(t, tc.expectedIndex, index)
		})
	}
}
