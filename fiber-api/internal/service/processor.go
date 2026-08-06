package service

import (
	"context"

	"fiber-api/internal/domain"
)

// StatisticsClient se define acá, no en client, para que Process no
// dependa de un cliente HTTP concreto y se pueda testear con un doble.
type StatisticsClient interface {
	Calculate(ctx context.Context, matrices map[string]domain.Matrix) (map[string]any, error)
}

// ProcessResult es la salida completa de Process.
type ProcessResult struct {
	Original   domain.Matrix
	Rotated    domain.Matrix // nil cuando no fue necesario rotar
	WasRotated bool
	Q          domain.Matrix
	R          domain.Matrix
	Statistics map[string]any
}

// MatrixProcessor implementa el caso de uso principal.
type MatrixProcessor struct {
	stats StatisticsClient
}

// NewMatrixProcessor arma el procesador con su cliente de estadísticas.
func NewMatrixProcessor(stats StatisticsClient) *MatrixProcessor {
	return &MatrixProcessor{stats: stats}
}

// Process valida, rota si la matriz es ancha, factoriza y pide las
// estadísticas.

// Si rota, la factorización corresponde a la matriz rotada, no a la
// original — el resultado lo declara vía WasRotated y Rotated.

func (p *MatrixProcessor) Process(ctx context.Context, input domain.Matrix) (*ProcessResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	result := &ProcessResult{Original: input}
	target := input

	if input.IsWide() {
		target = Rotate90Clockwise(input)
		result.Rotated = target
		result.WasRotated = true
	}

	qr, err := FactorizeQR(target)
	if err != nil {
		return nil, err
	}

	result.Q = qr.Q
	result.R = qr.R

	stats, err := p.stats.Calculate(ctx, map[string]domain.Matrix{
		"Q": qr.Q,
		"R": qr.R,
	})
	if err != nil {
		return nil, err
	}

	result.Statistics = stats

	return result, nil
}
