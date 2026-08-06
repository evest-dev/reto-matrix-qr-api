// Package domain no importa nada externo: ni Fiber, ni gonum, ni HTTP.
package domain

// Matrix es una matriz rectangular por convención: el compilador no lo
// exige, solo Validate lo comprueba.
type Matrix [][]float64

// Rows es la cantidad de filas.
func (m Matrix) Rows() int {
	return len(m)
}

// Cols es la cantidad de columnas; 0 para una matriz vacía.
func (m Matrix) Cols() int {
	if len(m) == 0 {
		return 0
	}
	return len(m[0])
}

// IsWide indica más columnas que filas: esa forma no admite factorización
// QR, hay que rotar la matriz antes.
func (m Matrix) IsWide() bool {
	return m.Rows() < m.Cols()
}

// Validate rechaza una matriz vacía o con filas de longitud distinta.
func (m Matrix) Validate() error {
	if len(m) == 0 {
		return ErrEmptyMatrix
	}

	width := len(m[0])
	if width == 0 {
		return ErrEmptyMatrix
	}

	for i, row := range m {
		if len(row) != width {
			return NewInconsistentRowError(i, width, len(row))
		}
	}

	return nil
}

// Flatten aplana la matriz por filas: es el orden que espera gonum.
func (m Matrix) Flatten() []float64 {
	out := make([]float64, 0, m.Rows()*m.Cols())
	for _, row := range m {
		out = append(out, row...)
	}
	return out
}
