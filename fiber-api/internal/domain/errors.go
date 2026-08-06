package domain

import (
	"errors"
	"fmt"
)

// Sentinel errors de dominio: las capas externas usan errors.Is para
// mapearlos a un código HTTP.
var (
	ErrEmptyMatrix     = errors.New("la matriz no puede estar vacía")
	ErrNotFactorizable = errors.New("la factorización QR requiere filas ≥ columnas")
)

// InconsistentRowError marca una fila de longitud distinta a las demás:
// la matriz no es rectangular.
type InconsistentRowError struct {
	RowIndex int
	Expected int
	Actual   int
}

func (e *InconsistentRowError) Error() string {
	return fmt.Sprintf(
		"la fila %d tiene %d columnas, se esperaban %d: la matriz no es rectangular",
		e.RowIndex, e.Actual, e.Expected,
	)
}

// NewInconsistentRowError arma el error con los tres datos de la fila inválida.
func NewInconsistentRowError(index, expected, actual int) error {
	return &InconsistentRowError{RowIndex: index, Expected: expected, Actual: actual}
}
