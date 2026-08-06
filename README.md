# Procesamiento de matrices — Factorización QR y estadísticas

Reto técnico · Interseguro, División TI

Dos APIs REST independientes que se comunican por HTTP. La primera recibe una matriz rectangular y calcula su factorización QR; la segunda calcula estadísticas descriptivas sobre las matrices resultantes.

---

## Arquitectura

```
Cliente
   │
   │ POST /api/v1/matrix/process
   ▼
┌──────────────────────────────────────┐
│  fiber-api · Go 1.26 · Fiber v2      │  :3000
│  valida → rota si hace falta →       │
│  factoriza QR                        │
└──────────────┬───────────────────────┘
               │ POST /api/v1/statistics
               ▼
┌──────────────────────────────────────┐
│  express-api · Node 20 · Express 5   │  :4000
│  max · min · sum · average ·         │
│  isAnyDiagonal                       │
└──────────────┬───────────────────────┘
               │
               └──► respuesta a fiber-api,
                    que compone el resultado final
```

El cliente solo interactúa con `fiber-api`. El servicio de estadísticas no se expone directamente.

---

## Despliegue

| Servicio | URL |
|---|---|
| API de factorización | https://fiber-api-htjz.onrender.com |
| API de estadísticas | https://express-api-v26w.onrender.com |

El cliente interactúa únicamente con la API de factorización. La de estadísticas se documenta por transparencia, pero no requiere consumo directo.

Prueba rápida del flujo completo:

```bash
curl -X POST https://fiber-api-htjz.onrender.com/api/v1/matrix/process \
  -H "Content-Type: application/json" \
  -d '{"matrix": [[0,5,0],[3,4,0]]}'
```

Ambos servicios corren en contenedores sobre el plan gratuito de Render. Tras quince minutos de inactividad se suspenden, por lo que la primera petición puede tardar cerca de un minuto mientras arrancan. Las siguientes responden con normalidad.

La ruta raíz `/` no está definida en ninguno de los dos servicios: los endpoints disponibles son los listados más abajo.

---

## Requisitos

**Con Docker** (recomendado): Docker y Docker Compose.

**Sin Docker**: Go 1.26 o superior, Node 20 o superior.

---

## Ejecución con Docker

La forma más rápida de levantar el sistema completo:

```bash
docker compose up --build
```

Ambos servicios quedan disponibles en `http://localhost:3000` (factorización) y `http://localhost:4000` (estadísticas). La comunicación entre ellos se resuelve por nombre de servicio en la red interna de Docker.

Para detenerlos:

```bash
docker compose down
```

Cada servicio tiene su propio `Dockerfile` con build multietapa: la imagen final contiene únicamente el binario compilado en el caso de Go, o el código transpilado en el de Node, sin herramientas de compilación ni dependencias de desarrollo, y se ejecuta con un usuario sin privilegios.

---

## Ejecución sin Docker

### fiber-api

```bash
cd fiber-api
go mod download
go run ./cmd/api
```

Disponible en `http://localhost:3000`.

### express-api

```bash
cd express-api
npm install
npm run dev
```

Disponible en `http://localhost:4000`.

Ambos servicios deben estar levantados para que el flujo completo funcione. `fiber-api` responde en `/rotate` sin necesidad del segundo servicio.

---

## Uso

Los ejemplos apuntan a los servicios desplegados. Para probar en local, reemplaza el host por `http://localhost:3000` en la API de factorización y `http://localhost:4000` en la de estadísticas.

### Procesar una matriz

```bash
curl -X POST https://fiber-api-htjz.onrender.com/api/v1/matrix/process \
  -H "Content-Type: application/json" \
  -d '{"matrix": [[3,0],[4,5],[0,0]]}'
```

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
    "max": 3,
    "min": -5,
    "sum": -6.6,
    "average": -0.44,
    "isAnyDiagonal": false
  }
}
```

### Procesar una matriz ancha

Cuando la matriz tiene más columnas que filas, se rota antes de factorizar. La respuesta lo declara explícitamente:

```bash
curl -X POST https://fiber-api-htjz.onrender.com/api/v1/matrix/process \
  -H "Content-Type: application/json" \
  -d '{"matrix": [[0,5,0],[3,4,0]]}'
```

```json
{
  "success": true,
  "original": [[0,5,0],[3,4,0]],
  "wasRotated": true,
  "rotated": [[3,0],[4,5],[0,0]],
  "factorizedFrom": "rotated",
  "qrFactorization": { "q": "...", "r": "..." },
  "statistics": { "...": "..." }
}
```

### Rotación aislada

```bash
curl -X POST https://fiber-api-htjz.onrender.com/api/v1/matrix/rotate \
  -H "Content-Type: application/json" \
  -d '{"matrix": [[1,2,3],[4,5,6]]}'
```

```json
{
  "original": [[1,2,3],[4,5,6]],
  "rotated": [[4,1],[5,2],[6,3]]
}
```

### Estadísticas directamente

```bash
curl -X POST https://express-api-v26w.onrender.com/api/v1/statistics \
  -H "Content-Type: application/json" \
  -d '{"matrices":[{"name":"D","data":[[5,0,0],[0,3,0],[0,0,9]]}]}'
```

```json
{
  "max": 9,
  "min": 0,
  "sum": 17,
  "average": 1.8888888888888888,
  "isAnyDiagonal": true
}
```

### Errores

```bash
curl -X POST https://fiber-api-htjz.onrender.com/api/v1/matrix/process \
  -H "Content-Type: application/json" \
  -d '{"matrix": [[1,2],[3]]}'
```

```json
{
  "success": false,
  "error": "la fila 1 tiene 1 columnas, se esperaban 2: la matriz no es rectangular"
}
```

---

## Endpoints

### fiber-api · puerto 3000

| Método | Ruta | Descripción |
|---|---|---|
| POST | `/api/v1/matrix/process` | Flujo completo: validar, rotar si corresponde, factorizar, obtener estadísticas |
| POST | `/api/v1/matrix/rotate` | Rotación 90° horario, sin factorizar |
| GET | `/health` | Estado del servicio |

### express-api · puerto 4000

| Método | Ruta | Descripción |
|---|---|---|
| POST | `/api/v1/statistics` | Calcula las cinco métricas sobre las matrices recibidas |
| GET | `/health` | Estado del servicio |

---

## Pruebas

```bash
cd fiber-api && go test ./... -race -cover
cd express-api && npm test -- --coverage
```

---

## Configuración

| Variable | Servicio | Por defecto | Uso |
|---|---|---|---|
| `PORT` | ambos | `3000` / `4000` | Puerto de escucha |
| `STATISTICS_API_URL` | fiber-api | `http://localhost:4000` | Base de la API de estadísticas |
| `HTTP_TIMEOUT_SECONDS` | fiber-api | `10` | Tiempo máximo de espera hacia express-api |
| `NODE_ENV` | express-api | `development` | Controla el detalle de los errores expuestos |

---

## Decisiones de interpretación del enunciado

El enunciado indica en su página 4 que, ante dudas, se espera que el candidato tome decisiones informadas y las sustente. Esta sección resume esas decisiones.

### Rotación y factorización QR

El enunciado menciona **rotación** en la descripción arquitectónica (pág. 2) y **factorización QR** en la funcionalidad requerida (pág. 3), sin explicitar la relación entre ambas.

La factorización QR está definida únicamente para matrices con `filas ≥ columnas`, mientras que el enunciado admite matrices rectangulares de cualquier forma. Dado que la rotación de 90° intercambia las dimensiones de una matriz, se adoptó como el mecanismo que habilita el procesamiento de matrices anchas:

```
si filas ≥ columnas  →  factorizar directamente
si filas <  columnas  →  rotar 90° horario, luego factorizar
```

Así se cumplen ambas menciones del enunciado sin descartar ninguna, y la rotación adquiere un propósito funcional en lugar de ser un paso decorativo.

Cuando se rota, `Q × R` reconstruye la matriz rotada y no la original. La respuesta lo declara en el campo `factorizedFrom`, de modo que la identidad sea verificable por el consumidor.

Sustento completo, con el recorrido matemático paso a paso, en [`docs/sustento-rotacion-qr.html`](docs/sustento-rotacion-qr.html).

### Otras decisiones

| Punto sin especificar | Decisión |
|---|---|
| Ángulo y sentido de la rotación | 90° horario, por ser la convención y producir el intercambio de dimensiones requerido |
| Estadísticas por matriz o sobre el conjunto | Sobre el conjunto unificado de Q y R, siguiendo la redacción en plural |
| Reporte de la verificación de diagonal | Un único booleano, siguiendo "verificar si alguna matriz es diagonal" |
| Destinatario de la respuesta de estadísticas | Responde a `fiber-api`, que compone la respuesta final para el cliente |

---

## Sobre los resultados de la factorización

La descomposición QR **no es única**: está determinada salvo el signo de cada columna de Q, siempre que la fila correspondiente de R lo compense.

`gonum` emplea reflexiones de Householder, numéricamente más estables que el método de Gram-Schmidt del recorrido manual. Ambas factorizaciones satisfacen `Q × R = A` pero producen signos distintos.

Por esa razón las pruebas verifican la reconstrucción del producto con tolerancia `1e-9`, nunca valores literales de Q y R.

También es esperable ver decimales largos como `14.000000000000002` donde matemáticamente correspondería `14`: es acumulación normal de error en aritmética de punto flotante. No se redondean, para no ocultar la precisión real del cálculo.

---

## Documentación

| Documento | Contenido |
|---|---|
| [`ARCHITECTURE.md`](ARCHITECTURE.md) | Arquitectura del sistema y decisiones de diseño |
| [`docs/requerimientos-generales.md`](docs/requerimientos-generales.md) | Requerimientos con criterios de aceptación y trazabilidad al enunciado |
| [`docs/sustento-rotacion-qr.html`](docs/sustento-rotacion-qr.html) | Recorrido matemático de la decisión de rotación condicional |
| [`fiber-api/ARCHITECTURE.md`](fiber-api/ARCHITECTURE.md) | Arquitectura interna del servicio de factorización |
| [`fiber-api/docs/requerimientos-paso-1.md`](fiber-api/docs/requerimientos-paso-1.md) | Requerimientos del servicio en Go |
| [`express-api/ARCHITECTURE.md`](express-api/ARCHITECTURE.md) | Arquitectura interna del servicio de estadísticas |
| [`express-api/docs/requerimientos-paso-2.md`](express-api/docs/requerimientos-paso-2.md) | Requerimientos del servicio en Node |

---

## Estructura del proyecto

```
.
├── README.md
├── ARCHITECTURE.md
├── docs/
│   ├── requerimientos-generales.md
│   └── sustento-rotacion-qr.html
│
├── fiber-api/                    Go · Fiber · gonum
│   ├── ARCHITECTURE.md
│   ├── docs/
│   ├── cmd/api/
│   └── internal/
│       ├── domain/
│       ├── service/
│       ├── client/
│       ├── handler/
│       └── config/
│
└── express-api/                  Node · TypeScript · Express
    ├── ARCHITECTURE.md
    ├── docs/
    ├── src/
    │   ├── domain/
    │   ├── service/
    │   ├── routes/
    │   ├── middleware/
    │   └── config/
    └── tests/
        ├── unit/
        └── integration/
```
