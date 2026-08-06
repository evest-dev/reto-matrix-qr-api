package service

import "fiber-api/internal/domain"

// Rotate90Clockwise convierte una matriz ancha (filas < columnas), inviable
// para la factorización QR, en una alta que sí la admite: el elemento en
// (i, j) de una matriz m×n pasa a (j, m-1-i) del resultado, de n×m.

func Rotate90Clockwise(m domain.Matrix) domain.Matrix {
	rows, cols := m.Rows(), m.Cols()

	rotated := make(domain.Matrix, cols)
	for i := range rotated {
		rotated[i] = make([]float64, rows)
	}

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			rotated[j][rows-1-i] = m[i][j]
		}
	}

	return rotated
}
