# Requerimientos — Paso 1 · API en Go (Fiber)

Reto técnico Interseguro · División TI

Requerimientos específicos de la API en Go. Los transversales a ambos servicios están en `../../docs/requerimientos-generales.md`.

La columna **fase** indica cuándo se aborda cada requerimiento: *local* durante el desarrollo, *cierre* una vez que ambas APIs funcionan en la máquina.

---

## Responsabilidad del servicio

Recibir una matriz rectangular, prepararla si su forma lo exige, calcular su factorización QR, delegar el cálculo estadístico a la API en Node.js, y componer la respuesta final para el cliente.

---

## RF-01 · Recepción y validación de la matriz

**Origen:** pág. 3 — "reciba como entrada un array de arrays de números que represente una matriz rectangular"
**Fase:** local
**Etapa del plan:** 1

Acepta una matriz rectangular expresada como arreglo de arreglos de números.

**Criterios de aceptación**

- Acepta matrices de cualquier dimensión m×n, no necesariamente cuadradas.
- Acepta valores enteros y decimales, positivos y negativos.
- Rechaza con HTTP 400 una petición cuyo cuerpo no sea JSON válido.
- Rechaza con HTTP 400 una matriz vacía.
- Rechaza con HTTP 400 una matriz cuyas filas tengan longitudes distintas, indicando el índice de la fila inconsistente.
- La validación reside en la capa de dominio, no en el handler.

**Verificación**

```bash
curl -X POST http://localhost:3000/api/v1/matrix/process \
  -H "Content-Type: application/json" \
  -d '{"matrix": [[1,2],[3]]}'
# 400 · indica que la fila 1 tiene 1 columna y se esperaban 2
```

---

## RF-02 · Rotación de la matriz

**Origen:** pág. 2 — "realizará la rotación de la matriz"
**Fase:** local
**Etapa del plan:** 2

Implementa la rotación de 90° en sentido horario, aplicada de forma condicional como paso habilitante de la factorización.

**Fundamento de la decisión.** La factorización QR exige `filas ≥ columnas`, mientras que el enunciado admite matrices rectangulares de cualquier forma. La rotación de 90° intercambia las dimensiones de la matriz (m×n → n×m), convirtiendo una matriz ancha inviable en una viable. Sustento completo en `../../sustento-rotacion-qr.html`.

**Criterios de aceptación**

- Transforma una matriz m×n en una n×m.
- El elemento en la posición (i, j) pasa a la posición (j, m−1−i).
- Se aplica automáticamente y **solo** cuando `filas < columnas`.
- **No** se aplica cuando la matriz ya satisface `filas ≥ columnas`: rotarla la volvería inviable.
- Se expone además como operación independiente en `POST /api/v1/matrix/rotate`.
- Opera en una sola pasada sobre los elementos.

**Verificación**

```bash
curl -X POST http://localhost:3000/api/v1/matrix/rotate \
  -H "Content-Type: application/json" \
  -d '{"matrix": [[0,5,0],[3,4,0]]}'
# rotated: [[3,0],[4,5],[0,0]]
```

---

## RF-03 · Factorización QR

**Origen:** pág. 3 — "devuelva la factorización QR de dicha matriz"
**Fase:** local
**Etapa del plan:** 3

Calcula la descomposición QR de la matriz efectiva, produciendo dos matrices Q y R cuyo producto la reconstruye.

**Criterios de aceptación**

- Devuelve Q ortonormal de dimensiones m×m.
- Devuelve R triangular superior de dimensiones m×n.
- Se cumple `Q × R = A` con tolerancia 1e-9, donde A es la matriz efectivamente factorizada.
- Todos los elementos de R por debajo de la diagonal principal son cero.
- La condición `filas ≥ columnas` se verifica antes de invocar a la librería, sin depender de su panic.
- La implementación delega en gonum; no se escribe un algoritmo propio.
- Los tipos de gonum no trascienden esta capa: se convierten a tipos del dominio antes de devolverse.

**Nota sobre la verificación.** Las pruebas comprueban la reconstrucción `Q × R = A`, nunca valores literales de Q y R. Gonum emplea reflexiones de Householder, numéricamente más estables que Gram-Schmidt, y los signos resultantes pueden diferir de un cálculo manual. Ambas factorizaciones son válidas.

**Valores de referencia** para la entrada `[[3,0],[4,5],[0,0]]`.

Salida real de gonum, mediante reflexiones de Householder:

```
Q = -0.6  -0.8  0       R = -5  -4
    -0.8   0.6  0            0   3
     0     0    1            0   0
```

El mismo cálculo por Gram-Schmidt produce los signos opuestos en la primera columna de Q y la primera fila de R:

```
Q =  0.6  -0.8  0       R =  5   4
     0.8   0.6  0            0   3
     0     0    1            0   0
```

Ambas son factorizaciones válidas: la descomposición QR está determinada salvo el signo de cada columna de Q, compensado por el de la fila correspondiente de R. Lo que se conserva en ambas es la reconstrucción `Q × R = A`, y es lo único que las pruebas verifican.

---

## RF-04 · Orquestación del procesamiento

**Origen:** derivado de la arquitectura descrita en pág. 2
**Fase:** local
**Etapa del plan:** 4

Coordina el flujo completo: validación, decisión de rotar, factorización y delegación estadística.

**Criterios de aceptación**

- La decisión de rotar se toma en una única comparación: `filas < columnas`.
- Registra si la rotación se aplicó y sobre qué matriz se factorizó.
- La dependencia hacia la API de estadísticas se declara como interfaz en esta capa, permitiendo sustituirla por un doble de prueba.
- Recibe `context.Context` como primer parámetro para propagar cancelación y tiempos límite.

---

## RF-05 · Consumo de la API de estadísticas

**Origen:** pág. 2 — "enviará los datos resultantes a la segunda API en Node.js"
**Fase:** local
**Etapa del plan:** 5

Transmite las matrices Q y R a la API en Node.js y recupera las estadísticas.

**Criterios de aceptación**

- Envía ambas matrices identificadas por nombre en una sola petición.
- La URL base se lee de la variable de entorno `STATISTICS_API_URL`, con `http://localhost:4000` por defecto.
- El tiempo máximo de espera se lee de `HTTP_TIMEOUT_SECONDS`, con 10 por defecto.
- Ante fallo de red o respuesta no exitosa, devuelve un error descriptivo sin exponer detalles internos.
- Los errores se envuelven con contexto mediante `fmt.Errorf("...: %w", err)`.

---

## RF-06 · Composición de la respuesta

**Origen:** derivado de la arquitectura descrita en pág. 2
**Fase:** local
**Etapa del plan:** 6

Integra todo el resultado del procesamiento en una única respuesta al cliente.

**Criterios de aceptación**

- Incluye la matriz original tal como fue recibida.
- Incluye la matriz rotada únicamente cuando la rotación se aplicó.
- Declara en `factorizedFrom` sobre qué matriz se calculó la factorización: `"original"` o `"rotated"`.
- Incluye las matrices Q y R.
- Incluye las cinco estadísticas devueltas por la API en Node.js.
- Campos JSON en `camelCase`.

**Forma de la respuesta**

```json
{
  "success": true,
  "original": [[3,0],[4,5],[0,0]],
  "wasRotated": false,
  "factorizedFrom": "original",
  "qrFactorization": { "q": [[...]], "r": [[...]] },
  "statistics": {
    "max": 5, "min": -0.8, "sum": 14.2,
    "average": 0.947, "isAnyDiagonal": false
  }
}
```

---

## RNF-01 · Arquitectura por capas

**Origen:** pág. 4 — "se espera coherencia en su estructura"
**Fase:** local

**Criterios de aceptación**

- Separación en `domain`, `service`, `client`, `handler` y `config`.
- Las dependencias apuntan hacia adentro: `domain` no importa Fiber, gonum ni paquetes de red.
- La traducción de errores de dominio a códigos HTTP ocurre en un único punto.
- Las dependencias se inyectan por constructor.

---

## RNF-02 · Manejo de errores

**Origen:** pág. 2 — "siguiendo las mejores prácticas de codificación"
**Fase:** local

**Criterios de aceptación**

- Errores de dominio expresados como sentinelas o tipos propios, identificables con `errors.Is` y `errors.As`.
- Errores envueltos con contexto mediante `%w`, preservando la cadena.
- Formato de error uniforme: `{ "success": false, "error": "mensaje en español" }`.
- Ningún error se ignora silenciosamente.

---

## RNF-03 · Documentación del código

**Origen:** pág. 2 — "Documentar el código de manera clara y concisa"
**Fase:** local

**Criterios de aceptación**

- Todo identificador exportado lleva comentario en formato GoDoc.
- Los comentarios explican el porqué de las decisiones, no lo que el código ya dice.
- La función de rotación documenta su propósito habilitante respecto de la factorización.

---

## RNF-04 · Pruebas

**Origen:** pág. 3 — funcionalidad opcional · pág. 3 — "de manera eficiente y correcta"
**Fase:** local

**Criterios de aceptación**

- Pruebas de tabla con subtests y ejecución en paralelo.
- La rotación se prueba con matriz ancha, alta y cuadrada.
- Existe una prueba que verifica que la rotación intercambia las dimensiones.
- La factorización se prueba verificando la reconstrucción `Q × R = A`.
- Existe una prueba que confirma que R es triangular superior.
- Existe una prueba que confirma el rechazo de matrices anchas en la factorización.
- Comparación de flotantes con tolerancia 1e-9; nunca igualdad exacta.
- Suite ejecutable con `go test ./... -race`.

---

## RNF-05 · Contenerización

**Origen:** pág. 2 — "Utilizar Docker para contenerizar las aplicaciones"
**Fase:** cierre

**Criterios de aceptación**

- `Dockerfile` con build multietapa.
- Binario compilado con `CGO_ENABLED=0` para producir un ejecutable estático.
- Imagen final mínima, con usuario sin privilegios.
- Healthcheck apuntando a `GET /health`.

---

## Trazabilidad

| Requerimiento | Origen | Fase | Etapa |
|---|---|---|---|
| RF-01 Recepción y validación | pág. 3 | local | 1 |
| RF-02 Rotación | pág. 2 | local | 2 |
| RF-03 Factorización QR | pág. 3 | local | 3 |
| RF-04 Orquestación | derivado | local | 4 |
| RF-05 Consumo de estadísticas | pág. 2 | local | 5 |
| RF-06 Composición de respuesta | derivado | local | 6 |
| RNF-01 Arquitectura por capas | pág. 4 | local | 1–6 |
| RNF-02 Manejo de errores | pág. 2 | local | 1–6 |
| RNF-03 Documentación | pág. 2 | local | 1–6 |
| RNF-04 Pruebas | pág. 3 | local | 2–3 |
| RNF-05 Contenerización | pág. 2 | cierre | — |

---

## Endpoints

| Método | Ruta | Descripción |
|---|---|---|
| POST | `/api/v1/matrix/process` | Flujo completo: validar, rotar si aplica, factorizar, estadísticas |
| POST | `/api/v1/matrix/rotate` | Solo rotación, sin factorizar |
| GET | `/health` | Estado del servicio, sin versionar |
