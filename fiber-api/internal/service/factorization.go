package service

import (
	"fiber-api/internal/domain"

	"gonum.org/v1/gonum/mat"
)

// QRResult reúne Q y R.
type QRResult struct {
	Q domain.Matrix // ortonormal, de dimensiones m×m
	R domain.Matrix // triangular superior, de dimensiones m×n
}

// FactorizeQR descompone m en Q y R tales que Q × R reconstruye m.
// Exige filas ≥ columnas: se comprueba antes de llamar a gonum, que
// entra en panic si no se cumple.

func FactorizeQR(m domain.Matrix) (*QRResult, error) {
	if m.IsWide() {
		return nil, domain.ErrNotFactorizable
	}

	dense := mat.NewDense(m.Rows(), m.Cols(), m.Flatten())

	var qr mat.QR
	qr.Factorize(dense)

	var q, r mat.Dense
	qr.QTo(&q)
	qr.RTo(&r)

	return &QRResult{
		Q: fromDense(&q),
		R: fromDense(&r),
	}, nil
}

// fromDense evita que el tipo de gonum se filtre hacia capas superiores.
func fromDense(d *mat.Dense) domain.Matrix {
	rows, cols := d.Dims()

	out := make(domain.Matrix, rows)
	for i := 0; i < rows; i++ {
		out[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			// -0 + 0 = +0 por IEEE 754: normaliza el cero negativo que
			// dejan las reflexiones de Householder de gonum.
			out[i][j] = d.At(i, j) + 0
		}
	}

	return out
}
