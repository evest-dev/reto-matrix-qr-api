package handler

import (
	"fiber-api/internal/domain"
	"fiber-api/internal/service"
)

// ProcessRequest es el cuerpo esperado en las peticiones de procesamiento.
type ProcessRequest struct {
	Matrix domain.Matrix `json:"matrix"`
}

// ProcessResponse es el body de POST /matrix/process.
type ProcessResponse struct {
	Success        bool           `json:"success"`
	Original       domain.Matrix  `json:"original"`
	WasRotated     bool           `json:"wasRotated"`
	Rotated        domain.Matrix  `json:"rotated,omitempty"`
	FactorizedFrom string         `json:"factorizedFrom"`
	QR             qrPayload      `json:"qrFactorization"`
	Statistics     map[string]any `json:"statistics"`
}

type qrPayload struct {
	Q domain.Matrix `json:"q"`
	R domain.Matrix `json:"r"`
}

// NewProcessResponse declara en FactorizedFrom qué matriz se factorizó,
// así el consumidor puede verificar Q × R sin ambigüedad.
func NewProcessResponse(r *service.ProcessResult) ProcessResponse {
	source := "original"
	if r.WasRotated {
		source = "rotated"
	}

	return ProcessResponse{
		Success:        true,
		Original:       r.Original,
		WasRotated:     r.WasRotated,
		Rotated:        r.Rotated,
		FactorizedFrom: source,
		QR:             qrPayload{Q: r.Q, R: r.R},
		Statistics:     r.Statistics,
	}
}
