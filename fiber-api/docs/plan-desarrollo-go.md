# Plan de desarrollo — API en Go, desde cero

Guía paso a paso asumiendo que no has escrito Go antes. Cada etapa explica qué vas a construir, qué conceptos del lenguaje aparecen, qué pedirle a Claude Code, y cómo verificar que quedó bien antes de avanzar.

**Principio rector:** construir de adentro hacia afuera, una capa a la vez, verificando cada una. No pedir todo de golpe — en la entrevista te van a preguntar por qué hiciste cada cosa.

---

## Etapa 0 · Preparar el entorno

### 0.1 Instalar Go

```bash
brew install go
go version
```

Debe imprimir algo como `go version go1.23.x darwin/arm64`.

### 0.2 Crear la estructura del monorepo

```bash
mkdir reto-interseguro && cd reto-interseguro
git init
mkdir -p fiber-api express-api docs
```

Coloca aquí los archivos que ya tienes: `CLAUDE.md` en la raíz, `fiber-api/CLAUDE.md`, y los documentos en `docs/`.

### 0.3 Inicializar el módulo de Go

```bash
cd fiber-api
go mod init fiber-api
```

**Qué acaba de pasar:** se creó un archivo `go.mod`. Es el equivalente al `package.json` de Node: declara el nombre del módulo y sus dependencias. El nombre que le des (`fiber-api`) es el prefijo que usarás en los imports internos: `import "fiber-api/internal/domain"`.

### 0.4 Instalar dependencias

```bash
go get github.com/gofiber/fiber/v2
go get gonum.org/v1/gonum/mat
```

Aparece un `go.sum` con los hashes de verificación. Ambos archivos se versionan en git.

---

## Conceptos de Go que vas a necesitar

Lee esto una vez antes de empezar. No necesitas dominarlo, solo reconocerlo cuando lo veas.

### Paquetes

Cada carpeta es un paquete. La primera línea de cada archivo declara a cuál pertenece:

```go
package domain
```

Todos los archivos de una carpeta deben declarar el mismo paquete. Para usar algo de otro paquete, lo importas y lo llamas con su prefijo: `domain.Matrix`.

### Exportado vs privado

No hay palabras clave `public` o `private`. **La mayúscula inicial decide:**

```go
func Rotate90Clockwise() {}  // exportada: visible desde otros paquetes
func normalize() {}          // privada: solo dentro de su paquete
```

Lo mismo aplica a tipos, campos de struct y variables.

### Tipos propios

Puedes ponerle nombre a un tipo existente y colgarle métodos:

```go
type Matrix [][]float64        // una matriz es una lista de listas de flotantes

func (m Matrix) Rows() int {   // método sobre ese tipo
    return len(m)
}
```

El `(m Matrix)` antes del nombre se llama *receptor*: es el equivalente a `this` en otros lenguajes.

### Errores

Go no tiene excepciones. Las funciones devuelven el error como último valor de retorno, y tú lo revisas:

```go
result, err := hacerAlgo()
if err != nil {
    return nil, err
}
```

Ese patrón `if err != nil` lo vas a ver por todas partes. Es normal, no es que el código esté mal escrito.

### Interfaces

Se definen por comportamiento y se cumplen implícitamente. No hay `implements`:

```go
type StatisticsClient interface {
    Calculate(ctx context.Context, m map[string]domain.Matrix) (map[string]any, error)
}
```

Cualquier tipo que tenga un método `Calculate` con esa firma exacta satisface la interfaz automáticamente. Esto es lo que te permite sustituir el cliente HTTP real por uno falso en los tests.

### Slices

`[]float64` es un slice: una lista de tamaño variable. Se crea con `make([]float64, n)` y se le agregan elementos con `append`.

### Tests

Van en archivos que terminan en `_test.go`, en la misma carpeta que el código que prueban. Cada test es una función que empieza con `Test`:

```go
func TestAlgo(t *testing.T) {
    if got != want {
        t.Errorf("got %v, want %v", got, want)
    }
}
```

Se ejecutan con `go test ./...`.

---

## Etapa 1 · El dominio

**Qué construyes:** la entidad `Matrix` con sus validaciones. Es el corazón del sistema y la capa que más importa entender, porque aquí vive la regla de `filas ≥ columnas`.

**Archivos:** `internal/domain/matrix.go`, `internal/domain/errors.go`

**Prompt para Claude Code:**

> Implementa `internal/domain/matrix.go` y `internal/domain/errors.go` según el diseño técnico en `docs/diseno-tecnico.md`, sección 4.1 y 4.2. Solo esos dos archivos. No agregues dependencias externas: el dominio no debe importar nada fuera de la librería estándar.

**Qué revisar antes de seguir:**

- ¿`Validate()` rechaza matriz vacía y filas de distinta longitud?
- ¿`IsWide()` devuelve `true` solo cuando hay más columnas que filas?
- ¿`Flatten()` recorre por filas, no por columnas? (gonum espera orden por filas)
- ¿Ningún import fuera de `errors` y `fmt`?

**Verificación:**

```bash
go build ./...
```

Debe compilar sin salida. En Go, "sin salida" significa éxito.

---

## Etapa 2 · La rotación

**Qué construyes:** la función que rota 90° horario. Es una función pura — misma entrada, misma salida, sin efectos secundarios — así que es fácil de verificar.

**Archivos:** `internal/service/rotation.go`, `internal/service/rotation_test.go`

**Prompt:**

> Implementa `internal/service/rotation.go` con la función `Rotate90Clockwise`, y su archivo de tests, según la sección 4.3 y 5.1 del diseño técnico. Los tests deben ser de tabla con `t.Parallel()`, e incluir el caso que verifica que la rotación intercambia las dimensiones.

**El detalle a entender:** la fórmula del reacomodo.

```go
rotated[j][rows-1-i] = m[i][j]
```

El elemento de la fila `i`, columna `j` termina en la fila `j`, columna `rows-1-i`. Verifícalo a mano con la matriz del ejemplo:

```
entrada 2×3        salida 3×2
 0  5  0            3  0
 3  4  0            4  5
                    0  0
```

El `3` estaba en `[1][0]` y termina en `[0][2-1-1] = [0][0]`. Correcto.

**Verificación:**

```bash
go test ./internal/service/ -run TestRotate -v
```

Los tres subtests deben pasar. Si entiendes esta función, entiendes la pieza central de tu decisión de diseño.

---

## Etapa 3 · La factorización

**Qué construyes:** el adaptador que envuelve a gonum. La matemática la hace la librería; tu trabajo es traducir entre tu tipo `Matrix` y el tipo de gonum, y validar antes de llamar.

**Archivos:** `internal/service/factorization.go`, `internal/service/factorization_test.go`

**Prompt:**

> Implementa `internal/service/factorization.go` y sus tests, según las secciones 4.4 y 5.2 del diseño técnico. Verifica la condición de forma antes de invocar a gonum: su implementación entra en panic ante matrices anchas y no queremos depender de eso.

**Lo importante de esta etapa:** el test principal **no** compara Q y R contra valores literales. Multiplica `Q × R` y verifica que reconstruya la entrada. La razón: gonum usa reflexiones de Householder, y los signos que produce pueden diferir de un cálculo manual por Gram-Schmidt. Ambos resultados son correctos. Si escribieras el test con valores fijos, fallaría sin que nada esté mal.

**Verificación:**

```bash
go test ./internal/service/ -v
```

Presta atención a `TestFactorizeQR_ReconstructsInput`. Si ese pasa, la matemática de tu API es correcta.

Prueba también, manualmente, factorizar la matriz del ejemplo e imprimir Q y R. Compáralos con los del documento de sustento: los valores absolutos deberían coincidir aunque los signos varíen.

---

## Etapa 4 · El orquestador

**Qué construyes:** el caso de uso que decide si rotar y encadena todo. Aquí se materializa tu decisión de diseño.

**Archivos:** `internal/service/processor.go`

**Prompt:**

> Implementa `internal/service/processor.go` según la sección 4.5 del diseño técnico. La interfaz `StatisticsClient` se define en esta capa, no en la del cliente, para poder sustituirla por un doble de prueba.

**El detalle conceptual:** fíjate que la interfaz `StatisticsClient` se declara aquí, en `service`, aunque la implementación real viva en `client`. Eso es inversión de dependencias: el caso de uso define qué necesita, y la capa externa se adapta. Es lo que permite testear el flujo completo sin levantar la API de Express.

**Verificación:** por ahora solo `go build ./...`. El test de esta capa viene después, cuando tengas el doble de prueba.

---

## Etapa 5 · El cliente HTTP

**Qué construyes:** la pieza que llama a la API de Express.

**Archivos:** `internal/client/statistics_client.go`, `internal/config/config.go`

**Prompt:**

> Implementa `internal/client/statistics_client.go` según la sección 4.6 del diseño técnico, y `internal/config/config.go` que lea las variables de entorno `PORT`, `STATISTICS_API_URL` y `HTTP_TIMEOUT_SECONDS` con valores por defecto.

**Importante:** la URL de la API de Express **debe** venir de variable de entorno, con `http://localhost:4000` como valor por defecto. Si la hardcodeas, cuando dockerices tendrás que cambiar código: dentro de Docker el servicio se llama `express-api`, no `localhost`.

**Verificación:** `go build ./...`. Aún no puedes probarlo de verdad porque la API de Express no existe.

---

## Etapa 6 · La capa HTTP

**Qué construyes:** los handlers de Fiber, los DTOs y el arranque de la aplicación.

**Archivos:** `internal/handler/matrix_handler.go`, `internal/handler/dto.go`, `cmd/api/main.go`

**Prompt:**

> Implementa `internal/handler/matrix_handler.go`, `internal/handler/dto.go` y `cmd/api/main.go` según las secciones 4.7, 4.8 y 4.9 del diseño técnico. La traducción de errores de dominio a códigos HTTP debe ocurrir en un único lugar.

**Verificación:**

```bash
go run ./cmd/api
```

En otra terminal:

```bash
curl http://localhost:3000/health
```

Debe responder `{"status":"ok"}`.

Prueba también la rotación, que no depende de la API de Express:

```bash
curl -X POST http://localhost:3000/api/v1/matrix/rotate \
  -H "Content-Type: application/json" \
  -d '{"matrix": [[0,5,0],[3,4,0]]}'
```

Debe devolver la matriz rotada `[[3,0],[4,5],[0,0]]`.

El endpoint `/process` fallará al llamar a Express — es esperado, todavía no existe.

---

## Etapa 7 · La API de Express

Recién ahora. Construir la segunda API cuando la primera ya funciona te permite probar el flujo completo de inmediato.

Es considerablemente más simple: recibir matrices, recorrer los valores, calcular cinco números.

**Verificación del flujo completo:**

```bash
curl -X POST http://localhost:3000/api/v1/matrix/process \
  -H "Content-Type: application/json" \
  -d '{"matrix": [[3,0],[4,5],[0,0]]}'
```

Debe devolver original, Q, R y las estadísticas. Contrasta los valores con el documento de sustento: máximo 5, mínimo −0.8, suma 14.2, promedio 0.947, diagonal `false`.

---

## Etapa 8 · Cierre

En este orden:

1. **Docker** — un `Dockerfile` por servicio y el `docker-compose.yml`. Como la configuración ya sale de variables de entorno, no hay que tocar código.
2. **README** — instalación, ejecución, ejemplos de petición y respuesta, y las decisiones de interpretación del enunciado.
3. **Despliegue en la nube** — requisito explícito del enunciado.
4. **JWT** — opcional. Si lo haces, protege los endpoints de procesamiento y deja `/health` público.
5. **Frontend** — opcional, el de menor peso relativo en la evaluación.

---

## Comandos de referencia

```bash
go run ./cmd/api              # ejecutar
go build ./...                # compilar todo
go test ./...                 # tests
go test ./... -race -cover    # tests con detector de carreras y cobertura
go test ./internal/service/ -v  # tests de un paquete, con detalle
go vet ./...                  # análisis estático
gofmt -w .                    # formatear (Go tiene un solo estilo, no se discute)
go mod tidy                   # limpiar dependencias no usadas
```

---

## Errores frecuentes al empezar con Go

**"declared and not used"** — Go no compila si declaras una variable y no la usas. Es un error, no una advertencia. Bórrala o úsala.

**"imported and not used"** — lo mismo con los imports.

**Los tests no corren** — el archivo debe terminar en `_test.go` y la función debe empezar con `Test` mayúscula y recibir `*testing.T`.

**"cannot find package"** — el import debe empezar con el nombre del módulo declarado en `go.mod`. Si tu módulo es `fiber-api`, el import es `fiber-api/internal/domain`, no `./internal/domain`.

**El formato cambia solo** — si tu editor tiene el plugin de Go, aplica `gofmt` al guardar. Es normal: Go tiene un único estilo oficial y no es configurable.

---

## Regla de trabajo

Después de cada etapa: lee el código generado, córrelo, y si algo no entiendes, pregunta antes de avanzar. El enunciado dice explícitamente que evaluarán tu capacidad de comunicar y defender tus decisiones técnicas. Código que no entiendes es código que no puedes defender.
