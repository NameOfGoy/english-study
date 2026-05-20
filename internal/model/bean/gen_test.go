package bean

import (
	"testing"

	"gorm.io/gen"
)

func TestGen(t *testing.T) {
	tests := []struct {
		name  string
		path  string
		model []interface{}
	}{
		{
			name:  "生成基本model",
			path:  "../dto",
			model: Schemas,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				g := gen.NewGenerator(
					gen.Config{
						OutPath: tt.path,
						Mode:    gen.WithoutContext | gen.WithQueryInterface, // generate mode
					},
				)
				g.ApplyBasic(
					tt.model...,
				)
				// GeneratePronounce the code
				g.Execute()
			},
		)
	}
}
