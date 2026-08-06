package service_test

import (
	"errors"
	"math"
	"testing"

	"fiber-api/internal/domain"
	"fiber-api/internal/service"
)

const tolerance = 1e-9

// TestFactorizeQR_ReconstructsInput exige Q × R == la matriz de entrada.
func TestFactorizeQR_ReconstructsInput(t *testing.T) {
	t.Parallel()

	input := domain.Matrix{{3, 0}, {4, 5}, {0, 0}}

	result, err := service.FactorizeQR(input)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	product := multiply(result.Q, result.R)

	for i := range input {
		for j := range input[i] {
			if math.Abs(product[i][j]-input[i][j]) > tolerance {
				t.Errorf("Q×R[%d][%d] = %v, se esperaba %v", i, j, product[i][j], input[i][j])
			}
		}
	}
}

// TestFactorizeQR_RejectsWideMatrix: gonum exige filas ≥ columnas.
func TestFactorizeQR_RejectsWideMatrix(t *testing.T) {
	t.Parallel()

	wide := domain.Matrix{{1, 2, 3}, {4, 5, 6}}

	_, err := service.FactorizeQR(wide)
	if !errors.Is(err, domain.ErrNotFactorizable) {
		t.Errorf("se esperaba ErrNotFactorizable, se obtuvo %v", err)
	}
}

// TestFactorizeQR_RIsUpperTriangular: nada bajo la diagonal principal.
func TestFactorizeQR_RIsUpperTriangular(t *testing.T) {
	t.Parallel()

	result, err := service.FactorizeQR(domain.Matrix{{3, 0}, {4, 5}, {0, 0}})
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	for i := range result.R {
		for j := range result.R[i] {
			if i > j && math.Abs(result.R[i][j]) > tolerance {
				t.Errorf("R[%d][%d] = %v, debería ser cero bajo la diagonal", i, j, result.R[i][j])
			}
		}
	}
}

func multiply(a, b domain.Matrix) domain.Matrix {
	rows, cols, inner := a.Rows(), b.Cols(), b.Rows()

	out := make(domain.Matrix, rows)
	for i := range out {
		out[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			for k := 0; k < inner; k++ {
				out[i][j] += a[i][k] * b[k][j]
			}
		}
	}

	return out
}
