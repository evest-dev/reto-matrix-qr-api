package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"fiber-api/internal/domain"
)

// StatisticsClient consume la API de estadísticas implementada en Node.
type StatisticsClient struct {
	baseURL string
	http    *http.Client
}

// NewStatisticsClient exige un timeout explícito: sin él, una demora de
// la API externa bloquearía indefinidamente.
func NewStatisticsClient(baseURL string, timeout time.Duration) *StatisticsClient {
	return &StatisticsClient{
		baseURL: baseURL,
		http:    &http.Client{Timeout: timeout},
	}
}

type matrixPayload struct {
	Name string        `json:"name"`
	Data domain.Matrix `json:"data"`
}

type statisticsRequest struct {
	Matrices []matrixPayload `json:"matrices"`
}

// Calculate llama a POST /api/v1/statistics.
func (c *StatisticsClient) Calculate(
	ctx context.Context,
	matrices map[string]domain.Matrix,
) (map[string]any, error) {

	payload := statisticsRequest{Matrices: make([]matrixPayload, 0, len(matrices))}
	for name, data := range matrices {
		payload.Matrices = append(payload.Matrices, matrixPayload{Name: name, Data: data})
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("serializando matrices: %w", err)
	}

	url := c.baseURL + "/api/v1/statistics"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("construyendo petición: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("consultando la API de estadísticas: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("la API de estadísticas respondió %d", resp.StatusCode)
	}

	var stats map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("decodificando estadísticas: %w", err)
	}

	return stats, nil
}
