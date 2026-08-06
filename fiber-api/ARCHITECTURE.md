# Arquitectura · fiber-api

Servicio en Go que recibe una matriz rectangular, la prepara si su forma lo exige, calcula su factorización QR y delega el cálculo estadístico al servicio en Node.

Las decisiones transversales al sistema están en [`../ARCHITECTURE.md`](../ARCHITECTURE.md).

---

## Stack

| Componente | Versión | Motivo |
|---|---|---|
| Go | 1.26 | — |
| Fiber | v2 | v3 se encontraba en beta al momento del desarrollo |
| gonum | `gonum.org/v1/gonum/mat` | Librería numérica establecida; evita implementar un algoritmo de ortogonalización propio |

---

## Estructura y regla de dependencia

```
cmd/api/main.go                     composición de dependencias y arranque
internal/domain/
  matrix.go                         entidad Matrix e invariantes
  errors.go                         errores de dominio tipados
internal/service/
  rotation.go                       rotación 90° horario
  factorization.go                  adaptador de gonum
  processor.go                      orquestación del caso de uso
internal/client/
  statistics_client.go              cliente HTTP hacia express-api
internal/handler/
  matrix_handler.go                 handlers Fiber
  dto.go                            objetos de transferencia
internal/config/
  config.go                         lectura del entorno
```

Las dependencias apuntan hacia adentro:

```
handler ──► service ──► domain
   │            │
   └────────────┴──► (domain no importa nada externo)

client ──► domain
```

`domain` no importa Fiber, gonum ni paquetes de red. Los tipos de gonum no trascienden `service/factorization.go`: se convierten a `domain.Matrix` antes de devolverse, de modo que un cambio de librería numérica no se propague al resto del sistema.

`internal/` no es una convención estética: el compilador de Go impide que código externo al módulo importe paquetes bajo ese directorio.

---

## Reglas de negocio

1. La factorización QR requiere `filas ≥ columnas`. Es una restricción matemática, no de la librería: `gonum` entra en panic ante una matriz ancha.
2. Si la matriz es ancha, se rota 90° horario **antes** de factorizar. La rotación intercambia dimensiones y la vuelve viable.
3. Si la matriz ya cumple la condición, **no se rota**: rotarla la volvería inviable.
4. Cuando se rota, `Q × R` reconstruye la matriz rotada, no la original. La respuesta lo declara en `factorizedFrom`.
5. La condición de forma se verifica en `FactorizeQR` antes de invocar a gonum, para no depender de un panic como mecanismo de control.

Fundamento completo de la decisión de rotación condicional en [`../ARCHITECTURE.md`](../ARCHITECTURE.md#2-la-decisión-de-diseño-central).

---

## Endpoints

| Método | Ruta | Descripción |
|---|---|---|
| POST | `/api/v1/matrix/process` | Flujo completo: validar, rotar si corresponde, factorizar, obtener estadísticas |
| POST | `/api/v1/matrix/rotate` | Rotación aislada, sin factorizar |
| GET | `/health` | Estado del servicio, sin versionar |

### Entrada

```json
{ "matrix": [[3, 0], [4, 5], [0, 0]] }
```

### Salida

```json
{
  "success": true,
  "original": [[3,0],[4,5],[0,0]],
  "wasRotated": false,
  "factorizedFrom": "original",
  "qrFactorization": {
    "q": [[-0.6,-0.8,0],[-0.8,0.6,0],[0,0,1]],
    "r": [[-5,-4],[0,3],[0,0]]
  },
  "statistics": {
    "max": 3, "min": -5, "sum": -6.6,
    "average": -0.44, "isAnyDiagonal": false
  }
}
```

Cuando la rotación se aplica, se agrega el campo `rotated` y `factorizedFrom` vale `"rotated"`.

---

## Manejo de errores

Los errores de dominio son sentinelas (`ErrEmptyMatrix`, `ErrNotFactorizable`) o tipos propios (`InconsistentRowError`), identificables con `errors.Is` y `errors.As`.

La traducción a códigos HTTP ocurre en un único punto, `mapDomainError` en `internal/handler`. Ninguna otra parte del código decide códigos de estado.

```json
{ "success": false, "error": "mensaje descriptivo" }
```

Los errores se envuelven con contexto mediante `fmt.Errorf("...: %w", err)`, preservando la cadena para diagnóstico.

---

## Ejecución

```bash
go run ./cmd/api              # levantar en local
go test ./... -race -cover    # pruebas con detector de carreras y cobertura
go vet ./...                  # análisis estático
gofmt -l .                    # listar archivos mal formateados
```

### Variables de entorno

| Variable | Valor por defecto | Uso |
|---|---|---|
| `PORT` | `3000` | Puerto de escucha |
| `STATISTICS_API_URL` | `http://localhost:4000` | Base de la API de estadísticas |
| `HTTP_TIMEOUT_SECONDS` | `10` | Tiempo máximo de espera hacia express-api |

---

## Convenciones

- Pruebas de tabla con subtests y ejecución en paralelo.
- Comparación de flotantes con tolerancia `1e-9`; nunca igualdad exacta.
- `context.Context` como primer parámetro en toda operación que realice entrada/salida.
- Todo identificador exportado documentado en formato GoDoc.
- Comentarios y mensajes de error en español; identificadores y campos JSON en inglés.

---

## Nota sobre las pruebas de factorización

La prueba central verifica que `Q × R` reconstruya la matriz de entrada, **no** valores literales de Q y R.

Gonum emplea reflexiones de Householder, numéricamente más estables que el método de Gram-Schmidt usado en el recorrido manual del documento de sustento. Ambos producen factorizaciones válidas, pero con distinta elección de signos. Una prueba escrita contra números fijos fallaría al cambiar de algoritmo o de versión de la librería sin que nada estuviera mal.

---

## Normalización del cero negativo

Las reflexiones de Householder producen `-0`, que es matemáticamente cero pero se serializa de forma distinta en JSON. Se normaliza en `fromDense`, el punto donde los datos de gonum se convierten al tipo del dominio:

```go
out[i][j] = d.At(i, j) + 0
```

IEEE 754 define que la suma de dos ceros de signo opuesto es cero positivo, por lo que `-0 + 0 = +0` mientras cualquier otro valor permanece inalterado. Normalizar en ese punto evita que el cero negativo alcance la respuesta JSON y la API de estadísticas.
