package service_test

import (
	"reflect"
	"testing"

	"fiber-api/internal/domain"
	"fiber-api/internal/service"
)

func TestRotate90Clockwise(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input domain.Matrix
		want  domain.Matrix
	}{
		{
			name:  "matriz ancha se vuelve alta",
			input: domain.Matrix{{0, 5, 0}, {3, 4, 0}},
			want:  domain.Matrix{{3, 0}, {4, 5}, {0, 0}},
		},
		{
			name:  "matriz alta se vuelve ancha",
			input: domain.Matrix{{3, 0}, {4, 5}, {0, 0}},
			want:  domain.Matrix{{0, 4, 3}, {0, 5, 0}},
		},
		{
			name:  "matriz cuadrada conserva dimensiones",
			input: domain.Matrix{{1, 2}, {3, 4}},
			want:  domain.Matrix{{3, 1}, {4, 2}},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := service.Rotate90Clockwise(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Rotate90Clockwise() = %v, se esperaba %v", got, tt.want)
			}
		})
	}
}

// TestRotationSwapsDimensions cubre la propiedad que justifica el diseño:
// intercambiar filas por columnas habilita la factorización.
func TestRotationSwapsDimensions(t *testing.T) {
	t.Parallel()

	wide := domain.Matrix{{1, 2, 3}, {4, 5, 6}}
	if !wide.IsWide() {
		t.Fatal("la matriz de prueba debería ser ancha")
	}

	rotated := service.Rotate90Clockwise(wide)

	if rotated.IsWide() {
		t.Error("tras rotar, la matriz no debería seguir siendo ancha")
	}
	if rotated.Rows() != wide.Cols() || rotated.Cols() != wide.Rows() {
		t.Errorf("dimensiones incorrectas: %dx%d", rotated.Rows(), rotated.Cols())
	}
}
