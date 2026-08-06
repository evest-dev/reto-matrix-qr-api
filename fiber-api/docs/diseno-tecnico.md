# API en Go (Fiber) — Diseño técnico y sustento

Reto técnico Interseguro · División TI

---

## 1. La decisión de diseño central

### 1.1 El problema

El enunciado presenta dos descripciones distintas de lo que debe hacer la API en Go:

| Sección | Texto | Operación |
|---|---|---|
| Arquitectura de la solución (pág. 2) | "realizará la rotación de la matriz" | Rotación |
| Funcionalidad requerida (pág. 3) | "devuelva la factorización QR de dicha matriz" | Factorización QR |

Son operaciones distintas. La rotación reordena los elementos existentes; la factorización QR produce dos matrices nuevas mediante cálculo.

### 1.2 La restricción que resuelve la ambigüedad

La factorización QR está definida únicamente para matrices con **filas ≥ columnas**. No es una limitación de implementación: es la condición bajo la cual la descomposición existe en la forma estándar.

Verificado en el código fuente de `gonum/mat` (`mat/qr.go`):

```go
// Factorize computes the QR factorization of an m×n matrix a where m >= n.
func (qr *QR) factorize(a Matrix, norm lapack.MatrixNorm) {
	m, n := a.Dims()
	if m < n {
		panic(ErrShape)
	}
	...
}
```

El enunciado, en cambio, admite explícitamente matrices **rectangulares** sin restricción de forma. Existe entonces un conjunto de entradas válidas según el enunciado (matrices anchas, con más columnas que filas) para las cuales la operación requerida no está definida.

### 1.3 La resolución adoptada

Rotar una matriz 90° **intercambia sus dimensiones**: una matriz de m×n se convierte en una de n×m. Una matriz 2×3, inviable para QR, se convierte en una 3×2 perfectamente viable.

Por lo tanto la rotación se adopta como **operación habilitante condicional**:

```
si filas ≥ columnas  →  factorizar directamente
si filas <  columnas  →  rotar 90°, luego factorizar
```

Esta decisión:

- Cumple ambas menciones del enunciado sin descartar ninguna.
- Le asigna a la rotación un propósito funcional, en vez de agregarla como paso decorativo.
- Preserva intacta la entrada del usuario en el caso mayoritario (matrices altas o cuadradas).
- Evita rechazar entradas que el enunciado declara válidas.

### 1.4 Alternativas evaluadas y descartadas

| Alternativa | Por qué se descartó |
|---|---|
| Rechazar matrices anchas con HTTP 400 | El enunciado las admite explícitamente al hablar de matrices rectangulares. Rechazarlas reduce el alcance solicitado. |
| Transponer en lugar de rotar | Matemáticamente equivalente en cuanto al cambio de dimensiones y algo más canónico, pero no está mencionada en el enunciado. La rotación sí lo está. |
| Aplicar factorización LQ a matrices anchas | Es la operación hermana correcta para el caso m < n, pero introduce una segunda descomposición no solicitada y complica el contrato de salida. |
| Rotar siempre, antes de todo | Rompe el caso que ya funcionaba: una matriz 3×2 rotada se convierte en 2×3, volviéndose inviable. La rotación debe ser condicional. |

### 1.5 Consecuencia declarada

Cuando se rota, la identidad que se cumple es:

```
Q × R = matriz rotada     (no la matriz original)
```

La respuesta de la API debe por tanto incluir la matriz rotada y señalar explícitamente que la factorización corresponde a ella. Así el consumidor puede verificar Q × R por su cuenta y obtener exactamente la matriz declarada.

---

## 2. Flujo de ejecución

```
                    ┌─────────────────────────┐
                    │  POST /matrix/process   │
                    │  { "matrix": [[...]] }  │
                    └───────────┬─────────────┘
                                │
                    ┌───────────▼─────────────┐
                    │  Validación de dominio  │
                    │  · no vacía             │
                    │  · filas homogéneas     │
                    │  · valores numéricos    │
                    └───────────┬─────────────┘
                                │
                    ┌───────────▼─────────────┐
                    │  ¿filas ≥ columnas?     │
                    └─────┬─────────────┬─────┘
                       sí │             │ no
                          │             │
                          │   ┌─────────▼──────────┐
                          │   │  Rotar 90° horario │
                          │   │  m×n  →  n×m       │
                          │   └─────────┬──────────┘
                          │             │
                    ┌─────▼─────────────▼─────┐
                    │  Factorización QR       │
                    │  gonum/mat              │
                    │  → Q (m×m), R (m×n)     │
                    └───────────┬─────────────┘
                                │
                    ┌───────────▼─────────────┐
                    │  HTTP POST → API Node   │
                    │  { matrices: [Q, R] }   │
                    └───────────┬─────────────┘
                                │
                    ┌───────────▼─────────────┐
                    │  API Node calcula:      │
                    │  max, min, avg,         │
                    │  sum, isAnyDiagonal     │
                    └───────────┬─────────────┘
                                │
                    ┌───────────▼─────────────┐
                    │  Respuesta compuesta    │
                    │  original + rotated? +  │
                    │  Q + R + statistics     │
                    └─────────────────────────┘
```

---

## 3. Estructura del proyecto

Aplicando Clean Architecture: las dependencias apuntan hacia adentro, el dominio no conoce ni a Fiber ni a gonum.

```
fiber-api/
├── cmd/
│   └── api/
│       └── main.go                  # composición y arranque
├── internal/
│   ├── domain/
│   │   ├── matrix.go                # entidad Matrix + invariantes
│   │   └── errors.go                # errores de dominio tipados
│   ├── service/
│   │   ├── rotation.go              # rotación 90°
│   │   ├── factorization.go         # QR (adaptador de gonum)
│   │   └── processor.go             # orquestación del caso de uso
│   ├── client/
│   │   └── statistics_client.go     # cliente HTTP hacia la API Node
│   ├── handler/
│   │   ├── matrix_handler.go        # handlers Fiber
│   │   └── dto.go                   # request/response
│   └── config/
│       └── config.go                # variables de entorno
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── go.sum
```

**Regla de dependencia:** `domain` no importa nada de `service`, `handler` ni librerías externas. `service` importa `domain`. `handler` importa `service` y `domain`. Esto permite testear la lógica de negocio sin levantar un servidor HTTP.

---

## 4. Código

### 4.1 Dominio — `internal/domain/matrix.go`

```go
// Package domain contiene las entidades y reglas de negocio del sistema.
// No depende de frameworks web, librerías numéricas ni detalles de transporte.
package domain

// Matrix representa una matriz rectangular de números reales,
// organizada como una lista de filas.
type Matrix [][]float64

// Rows devuelve la cantidad de filas de la matriz.
func (m Matrix) Rows() int {
	return len(m)
}

// Cols devuelve la cantidad de columnas de la matriz.
// Retorna 0 para una matriz vacía.
func (m Matrix) Cols() int {
	if len(m) == 0 {
		return 0
	}
	return len(m[0])
}

// IsWide informa si la matriz tiene más columnas que filas.
// Una matriz ancha no admite factorización QR sin transformación previa.
func (m Matrix) IsWide() bool {
	return m.Rows() < m.Cols()
}

// Validate verifica los invariantes de una matriz rectangular:
// que no esté vacía y que todas sus filas tengan la misma longitud.
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

// Flatten devuelve los valores de la matriz en un único slice,
// recorridos por filas. Es el formato que espera gonum.
func (m Matrix) Flatten() []float64 {
	out := make([]float64, 0, m.Rows()*m.Cols())
	for _, row := range m {
		out = append(out, row...)
	}
	return out
}
```

### 4.2 Errores de dominio — `internal/domain/errors.go`

```go
package domain

import (
	"errors"
	"fmt"
)

// Errores sentinela del dominio. Las capas externas los identifican
// con errors.Is para decidir el código HTTP correspondiente.
var (
	ErrEmptyMatrix     = errors.New("la matriz no puede estar vacía")
	ErrNotFactorizable = errors.New("la factorización QR requiere filas ≥ columnas")
)

// InconsistentRowError indica que una fila no tiene la longitud esperada,
// es decir, que la matriz no es rectangular.
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

// NewInconsistentRowError construye el error de fila inconsistente.
func NewInconsistentRowError(index, expected, actual int) error {
	return &InconsistentRowError{RowIndex: index, Expected: expected, Actual: actual}
}
```

### 4.3 Rotación — `internal/service/rotation.go`

```go
package service

import "fiber-api/internal/domain"

// Rotate90Clockwise rota la matriz 90 grados en sentido horario.
//
// El elemento en la posición (i, j) de una matriz de m×n pasa a la
// posición (j, m-1-i) del resultado, que tiene dimensiones n×m.
//
// El intercambio de dimensiones es la propiedad relevante para este
// sistema: convierte una matriz ancha (filas < columnas), inviable para
// la factorización QR, en una matriz alta que sí la admite.
//
//	entrada 2×3        salida 3×2
//	 0  5  0            3  0
//	 3  4  0            4  5
//	                    0  0
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
```

### 4.4 Factorización — `internal/service/factorization.go`

```go
package service

import (
	"fiber-api/internal/domain"

	"gonum.org/v1/gonum/mat"
)

// QRResult contiene las dos matrices producidas por la factorización.
type QRResult struct {
	Q domain.Matrix // ortonormal, de dimensiones m×m
	R domain.Matrix // triangular superior, de dimensiones m×n
}

// FactorizeQR calcula la factorización QR de la matriz recibida,
// tal que Q × R reconstruye exactamente la matriz de entrada.
//
// Requiere filas ≥ columnas. Esta condición se verifica antes de invocar
// a gonum, cuya implementación entra en panic ante una matriz ancha.
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

// fromDense convierte una matriz de gonum al tipo del dominio,
// evitando que el tipo de la librería se filtre hacia capas superiores.
func fromDense(d *mat.Dense) domain.Matrix {
	rows, cols := d.Dims()

	out := make(domain.Matrix, rows)
	for i := 0; i < rows; i++ {
		out[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			out[i][j] = d.At(i, j)
		}
	}

	return out
}
```

### 4.5 Orquestación — `internal/service/processor.go`

```go
package service

import (
	"context"

	"fiber-api/internal/domain"
)

// StatisticsClient abstrae el consumo de la API de estadísticas.
// Se define como interfaz en esta capa para poder sustituirla por un
// doble de prueba, sin acoplar el caso de uso a un cliente HTTP concreto.
type StatisticsClient interface {
	Calculate(ctx context.Context, matrices map[string]domain.Matrix) (map[string]any, error)
}

// ProcessResult reúne todo lo producido durante el procesamiento.
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

// NewMatrixProcessor construye el procesador con sus dependencias.
func NewMatrixProcessor(stats StatisticsClient) *MatrixProcessor {
	return &MatrixProcessor{stats: stats}
}

// Process ejecuta el flujo completo: valida la entrada, rota si la forma
// lo exige, factoriza, y delega el cálculo estadístico a la API externa.
//
// La rotación se aplica únicamente cuando la matriz es ancha. En ese caso
// la factorización corresponde a la matriz rotada, no a la original, y el
// resultado lo declara mediante WasRotated y Rotated.
func (p *MatrixProcessor) Process(ctx context.Context, input domain.Matrix) (*ProcessResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	result := &ProcessResult{Original: input}

	// La matriz que efectivamente se factoriza.
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
```

### 4.6 Cliente HTTP — `internal/client/statistics_client.go`

```go
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

// NewStatisticsClient construye el cliente con un timeout explícito,
// para que una demora de la API externa no bloquee indefinidamente.
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

// Calculate envía las matrices a la API de estadísticas y devuelve el resultado.
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
```

### 4.7 Handler Fiber — `internal/handler/matrix_handler.go`

```go
package handler

import (
	"errors"

	"fiber-api/internal/domain"
	"fiber-api/internal/service"

	"github.com/gofiber/fiber/v2"
)

// MatrixHandler expone el caso de uso a través de HTTP.
type MatrixHandler struct {
	processor *service.MatrixProcessor
}

// NewMatrixHandler construye el handler con su dependencia.
func NewMatrixHandler(p *service.MatrixProcessor) *MatrixHandler {
	return &MatrixHandler{processor: p}
}

// Register monta las rutas del recurso sobre el router recibido.
func (h *MatrixHandler) Register(router fiber.Router) {
	group := router.Group("/matrix")
	group.Post("/process", h.Process)
	group.Post("/rotate", h.Rotate)
}

// Process recibe una matriz, la procesa y devuelve el resultado completo.
func (h *MatrixHandler) Process(c *fiber.Ctx) error {
	var req ProcessRequest
	if err := c.BodyParser(&req); err != nil {
		return respondError(c, fiber.StatusBadRequest, "el cuerpo de la petición no es un JSON válido")
	}

	result, err := h.processor.Process(c.UserContext(), req.Matrix)
	if err != nil {
		return mapDomainError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(NewProcessResponse(result))
}

// Rotate expone la rotación de forma aislada, sin factorizar.
func (h *MatrixHandler) Rotate(c *fiber.Ctx) error {
	var req ProcessRequest
	if err := c.BodyParser(&req); err != nil {
		return respondError(c, fiber.StatusBadRequest, "el cuerpo de la petición no es un JSON válido")
	}

	if err := req.Matrix.Validate(); err != nil {
		return mapDomainError(c, err)
	}

	rotated := service.Rotate90Clockwise(req.Matrix)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"original": req.Matrix,
		"rotated":  rotated,
	})
}

// mapDomainError traduce errores de dominio al código HTTP correspondiente.
// Es el único punto donde el dominio se conecta con la semántica de HTTP.
func mapDomainError(c *fiber.Ctx, err error) error {
	var inconsistent *domain.InconsistentRowError

	switch {
	case errors.Is(err, domain.ErrEmptyMatrix),
		errors.Is(err, domain.ErrNotFactorizable),
		errors.As(err, &inconsistent):
		return respondError(c, fiber.StatusBadRequest, err.Error())
	default:
		return respondError(c, fiber.StatusInternalServerError, "error procesando la matriz")
	}
}

func respondError(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"success": false,
		"error":   message,
	})
}
```

### 4.8 DTOs — `internal/handler/dto.go`

```go
package handler

import (
	"fiber-api/internal/domain"
	"fiber-api/internal/service"
)

// ProcessRequest es el cuerpo esperado en las peticiones de procesamiento.
type ProcessRequest struct {
	Matrix domain.Matrix `json:"matrix"`
}

// ProcessResponse es la respuesta que se devuelve al cliente.
type ProcessResponse struct {
	Success    bool           `json:"success"`
	Original   domain.Matrix  `json:"original"`
	WasRotated bool           `json:"wasRotated"`
	Rotated    domain.Matrix  `json:"rotated,omitempty"`
	FactorizedFrom string     `json:"factorizedFrom"`
	QR         qrPayload      `json:"qrFactorization"`
	Statistics map[string]any `json:"statistics"`
}

type qrPayload struct {
	Q domain.Matrix `json:"q"`
	R domain.Matrix `json:"r"`
}

// NewProcessResponse arma la respuesta declarando explícitamente sobre qué
// matriz se calculó la factorización, para que el consumidor pueda
// verificar Q × R sin ambigüedad.
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
```

### 4.9 Arranque — `cmd/api/main.go`

```go
// Command api levanta el servicio HTTP de procesamiento de matrices.
package main

import (
	"log"
	"time"

	"fiber-api/internal/client"
	"fiber-api/internal/config"
	"fiber-api/internal/handler"
	"fiber-api/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	cfg := config.Load()

	app := fiber.New(fiber.Config{
		AppName:      "matrix-factorization-api",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Composición de dependencias: se construyen de adentro hacia afuera.
	statsClient := client.NewStatisticsClient(cfg.StatisticsAPIURL, cfg.HTTPTimeout)
	processor := service.NewMatrixProcessor(statsClient)
	matrixHandler := handler.NewMatrixHandler(processor)

	v1 := app.Group("/api/v1")
	matrixHandler.Register(v1)

	log.Printf("servicio escuchando en :%s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("no se pudo iniciar el servidor: %v", err)
	}
}
```

---

## 5. Pruebas

### 5.1 Rotación — `internal/service/rotation_test.go`

```go
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

// TestRotationSwapsDimensions verifica la propiedad que justifica el diseño:
// la rotación intercambia filas por columnas, habilitando la factorización.
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
```

### 5.2 Factorización — `internal/service/factorization_test.go`

```go
package service_test

import (
	"errors"
	"math"
	"testing"

	"fiber-api/internal/domain"
	"fiber-api/internal/service"
)

const tolerance = 1e-9

// TestFactorizeQR_ReconstructsInput verifica la propiedad fundamental:
// Q × R debe reconstruir la matriz de entrada.
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

// TestFactorizeQR_RejectsWideMatrix documenta la restricción de forma.
func TestFactorizeQR_RejectsWideMatrix(t *testing.T) {
	t.Parallel()

	wide := domain.Matrix{{1, 2, 3}, {4, 5, 6}}

	_, err := service.FactorizeQR(wide)
	if !errors.Is(err, domain.ErrNotFactorizable) {
		t.Errorf("se esperaba ErrNotFactorizable, se obtuvo %v", err)
	}
}

// TestFactorizeQR_RIsUpperTriangular verifica que R cumpla su forma.
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
```

Ejecución:

```bash
go test ./... -race -cover
```

---

## 6. Contenedor

`Dockerfile` con build multietapa, imagen final mínima y usuario no privilegiado:

```dockerfile
# ---- etapa de compilación ----
FROM golang:1.23-alpine AS builder

WORKDIR /build

# Las dependencias se copian primero para aprovechar la caché de capas:
# solo se reinstalan si go.mod o go.sum cambian.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO deshabilitado para producir un binario estático.
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /build/api ./cmd/api

# ---- etapa final ----
FROM alpine:3.20

RUN apk add --no-cache ca-certificates && \
    adduser -D -u 10001 appuser

COPY --from=builder /build/api /usr/local/bin/api

USER appuser

EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s \
    CMD wget -qO- http://localhost:3000/health || exit 1

ENTRYPOINT ["/usr/local/bin/api"]
```

`docker-compose.yml` con las dos APIs en una red común, resolviéndose por nombre de servicio:

```yaml
services:
  fiber-api:
    build: ./fiber-api
    ports:
      - "3000:3000"
    environment:
      PORT: "3000"
      STATISTICS_API_URL: "http://express-api:4000"
    depends_on:
      express-api:
        condition: service_healthy
    networks:
      - matrix-net

  express-api:
    build: ./express-api
    ports:
      - "4000:4000"
    environment:
      PORT: "4000"
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://localhost:4000/health"]
      interval: 30s
      timeout: 3s
      start_period: 5s
    networks:
      - matrix-net

networks:
  matrix-net:
    driver: bridge
```

---

## 7. Puntos de defensa para la entrevista

1. **La ambigüedad del enunciado se detectó y se resolvió con criterio.** No se descartó ninguna de las dos menciones; se encontró la relación funcional entre ambas.

2. **La restricción de forma se verificó en la fuente,** no se asumió. El código de `gonum/mat` confirma que la factorización exige filas ≥ columnas.

3. **La rotación tiene un propósito.** Es el mecanismo que habilita el procesamiento de matrices anchas, no un paso agregado para cumplir con una frase del enunciado.

4. **La transformación se declara explícitamente.** El campo `factorizedFrom` en la respuesta indica sobre qué matriz se calculó Q y R, de modo que la identidad Q × R sea verificable por el consumidor.

5. **Se evaluaron alternativas y se documentó por qué se descartaron:** rechazar la entrada, transponer, o usar factorización LQ.

6. **Los signos de Q y R pueden diferir según el algoritmo.** Gonum emplea reflexiones de Householder, numéricamente más estable que Gram-Schmidt; ambas producen factorizaciones válidas que satisfacen Q × R = A, aunque con signos distintos. Por eso las pruebas verifican la reconstrucción del producto y no valores literales.

7. **La arquitectura permite testear sin infraestructura.** El dominio no conoce Fiber ni gonum; el cliente de estadísticas es una interfaz sustituible por un doble de prueba.
