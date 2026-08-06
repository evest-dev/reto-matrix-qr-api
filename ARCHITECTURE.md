# Arquitectura del sistema

Sistema compuesto por dos APIs REST independientes que se comunican por HTTP para procesar matrices.

---

## 1. Visión general

```
Cliente
   │
   │ POST /api/v1/matrix/process
   ▼
┌──────────────────────────────────────┐
│  fiber-api · Go 1.26 · Fiber v2      │  :3000
│                                      │
│  valida → rota si hace falta →       │
│  factoriza QR                        │
└──────────────┬───────────────────────┘
               │ POST /api/v1/statistics
               ▼
┌──────────────────────────────────────┐
│  express-api · Node 20 · Express 5   │  :4000
│                                      │
│  max · min · sum · average ·         │
│  isAnyDiagonal                       │
└──────────────┬───────────────────────┘
               │
               └──► respuesta a fiber-api,
                    que compone el resultado final
```

El cliente nunca invoca a `express-api` directamente. `fiber-api` actúa como su cliente HTTP y compone la respuesta que devuelve al consumidor.

| Servicio | Carpeta | Stack | Puerto | Responsabilidad |
|---|---|---|---|---|
| Factorización | `fiber-api/` | Go 1.26 · Fiber v2 · gonum | 3000 | Validar, rotar cuando corresponde, factorizar en QR |
| Estadísticas | `express-api/` | Node 20 · TypeScript 5 · Express 5 | 4000 | Calcular cinco métricas sobre las matrices recibidas |

---

## 2. La decisión de diseño central

### 2.1 El problema

El enunciado presenta dos descripciones distintas de lo que debe hacer la API en Go:

| Sección | Texto | Operación |
|---|---|---|
| Arquitectura de la solución (pág. 2) | "realizará la rotación de la matriz" | Rotación |
| Funcionalidad requerida (pág. 3) | "devuelva la factorización QR de dicha matriz" | Factorización QR |

Son operaciones distintas: la rotación reordena los elementos existentes, la factorización produce dos matrices nuevas mediante cálculo.

### 2.2 La restricción que las vincula

La factorización QR está definida únicamente para matrices con **filas ≥ columnas**. No es una limitación de implementación, sino la condición bajo la cual la descomposición existe en su forma estándar. Verificado en el código fuente de `gonum/mat`:

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

El enunciado, en cambio, admite matrices **rectangulares** sin restricción de forma. Existe entonces un conjunto de entradas válidas según el enunciado —matrices con más columnas que filas— para las cuales la operación requerida no está definida.

### 2.3 La resolución adoptada

Rotar una matriz 90° **intercambia sus dimensiones**: una de m×n se convierte en una de n×m. Una matriz 2×3, inviable para QR, se convierte en una 3×2 perfectamente viable.

La rotación se adopta por tanto como **operación habilitante condicional**:

```
si filas ≥ columnas  →  factorizar directamente
si filas <  columnas  →  rotar 90° horario, luego factorizar
```

Esta decisión cumple ambas menciones del enunciado sin descartar ninguna, le asigna a la rotación un propósito funcional en lugar de agregarla como paso decorativo, preserva intacta la entrada del usuario en el caso mayoritario, y evita rechazar entradas que el enunciado declara válidas.

### 2.4 Alternativas evaluadas y descartadas

| Alternativa | Motivo del descarte |
|---|---|
| Rechazar matrices anchas con HTTP 400 | El enunciado las admite explícitamente al hablar de matrices rectangulares |
| Transponer en lugar de rotar | Equivalente en cuanto al cambio de dimensiones y algo más canónico, pero no está mencionada en el enunciado |
| Aplicar factorización LQ a matrices anchas | Es la operación hermana correcta para el caso `filas < columnas`, pero introduce una descomposición no solicitada y complica el contrato de salida |
| Rotar siempre, antes de todo | Rompe el caso que ya funcionaba: una matriz 3×2 rotada se convierte en 2×3, volviéndose inviable |

### 2.5 Consecuencia declarada

Cuando se rota, la identidad que se cumple es:

```
Q × R = matriz rotada     (no la matriz original)
```

Por eso la respuesta incluye la matriz rotada y declara en `factorizedFrom` sobre cuál se calculó la factorización. El consumidor puede verificar `Q × R` por su cuenta y obtener exactamente la matriz declarada.

---

## 3. Otras decisiones de interpretación

| Punto sin especificar en el enunciado | Decisión adoptada |
|---|---|
| Ángulo y sentido de la rotación | 90° horario, por ser la convención y producir el intercambio de dimensiones requerido |
| Estadísticas por matriz o sobre el conjunto | Sobre el conjunto unificado de Q y R, siguiendo la redacción en plural: "todos los valores de las matrices" |
| Forma de reportar la verificación de diagonal | Un único booleano, siguiendo "verificar si alguna matriz es diagonal" |
| Destinatario de la respuesta de la API de estadísticas | Responde a `fiber-api`, que compone la respuesta final para el cliente |
| Criterio de matriz diagonal en matrices no cuadradas | Se aplica el mismo criterio: todo elemento fuera de la diagonal principal debe ser cero |

---

## 4. Arquitectura interna

Ambos servicios aplican la misma separación en capas, adaptada a las convenciones de su lenguaje. Las dependencias apuntan hacia adentro: la capa de dominio no conoce frameworks web, librerías numéricas ni detalles de transporte.

### fiber-api

```
cmd/api/main.go                     composición de dependencias y arranque
internal/domain/                    entidades e invariantes · sin dependencias externas
internal/service/                   casos de uso · rotación, factorización, orquestación
internal/client/                    cliente HTTP hacia express-api
internal/handler/                   handlers Fiber y objetos de transferencia
internal/config/                    lectura del entorno
```

`internal/` no es una convención estética: el compilador de Go impide que código externo al módulo importe paquetes bajo ese directorio.

### express-api

```
src/domain/                         entidades e invariantes · sin dependencias externas
src/service/                        cálculo estadístico · funciones puras
src/routes/                         rutas Express, controladores y esquemas de validación
src/middleware/                     transversales: errores, validación, registro
src/config/                         lectura del entorno
src/app.ts                          composición de la aplicación, sin escuchar
src/server.ts                       arranque del servidor
```

`app.ts` exporta la aplicación sin invocar `listen`; `server.ts` la levanta. Esa separación permite ejercitar los endpoints con Supertest sin abrir un puerto real.

---

## 5. Decisiones técnicas transversales

**Configuración externalizada desde el inicio.** La URL de `express-api` se lee de la variable `STATISTICS_API_URL`, con `http://localhost:4000` por defecto. Dentro de Docker el servicio se resuelve por nombre y no por `localhost`; leer la URL del entorno permitió contenerizar sin modificar código.

**Delegación en librería numérica.** La factorización usa `gonum`, no una implementación propia. Es código numérico probado y mantenido; escribir un algoritmo de ortogonalización propio habría añadido superficie de error sin beneficio.

**Validación en dos capas en `express-api`.** Zod valida la forma del JSON en el borde; el dominio valida la coherencia matemática (filas homogéneas, valores finitos). Son problemas distintos y fallan con mensajes distintos.

**Traducción de errores en un único punto.** En Go, `mapDomainError`; en Node, el middleware `errorHandler`. Ninguna otra parte del código decide códigos de estado HTTP.

**Formato de error homogéneo entre servicios.** Ambos responden `{ "success": false, "error": "mensaje" }`.

**Respuesta de estadísticas sin envoltura.** `express-api` devuelve el objeto de métricas directo. `fiber-api` lo reenvía tal cual bajo la clave `statistics`; envolverlo obligaría a Go a desenvolverlo, añadiendo acoplamiento sin beneficio.

**Valores tratados como flotantes de doble precisión.** La factorización QR produce decimales y negativos aunque la matriz de entrada sea de enteros positivos.

---

## 6. Sobre la no unicidad de la factorización QR

La descomposición QR **no es única**: está determinada salvo el signo de cada columna de Q, siempre que la fila correspondiente de R lo compense. Distintos algoritmos eligen distintos signos y todos satisfacen la identidad fundamental.

Para la entrada `[[3,0],[4,5],[0,0]]`:

```
Gram-Schmidt                        Householder · lo que produce gonum

Q =  0.6  -0.8  0                   Q = -0.6  -0.8  0
     0.8   0.6  0                       -0.8   0.6  0
     0     0    1                        0     0    1

R =  5  4                           R = -5  -4
     0  3                                0   3
     0  0                                0   0
```

Ambas reconstruyen la matriz original:

```
(−0.6)(−5) + (−0.8)(0) + (0)(0) = 3
(−0.6)(−4) + (−0.8)(3) + (0)(0) = 2.4 − 2.4 = 0
(−0.8)(−5) + ( 0.6)(0) + (0)(0) = 4
(−0.8)(−4) + ( 0.6)(3) + (0)(0) = 3.2 + 1.8 = 5
```

**Consecuencia para las pruebas.** Verifican la reconstrucción `Q × R = A` con tolerancia `1e-9`, nunca valores literales de Q y R. Una prueba escrita contra números fijos fallaría al cambiar de algoritmo o de versión de la librería, sin que nada estuviera mal.

Recorrido matemático completo en [`docs/sustento-rotacion-qr.md`](docs/sustento-rotacion-qr.md).

---

## 7. Contenerización y despliegue

### Imágenes

Cada servicio tiene su propio `Dockerfile` con build multietapa. La imagen final contiene únicamente el artefacto ejecutable —el binario compilado en Go, el código transpilado en Node— sin compiladores, herramientas de construcción ni dependencias de desarrollo. Ambas se ejecutan con un usuario sin privilegios.

En Go el binario se compila con `CGO_ENABLED=0` para obtener un ejecutable estático que corra sobre una imagen base mínima.

### Orquestación local

`docker-compose.yml` levanta ambos servicios en una red común. La resolución entre ellos ocurre por nombre de servicio, no por dirección IP, y `fiber-api` espera a que `express-api` reporte estado saludable antes de arrancar.

### Configuración entre entornos

La misma variable toma tres valores distintos sin que el código cambie:

| Entorno | `STATISTICS_API_URL` |
|---|---|
| Local | `http://localhost:4000` |
| Docker Compose | `http://express-api:4000` |
| Producción | URL pública del servicio desplegado |

En local proviene del valor por defecto en la configuración; en Docker, del archivo de composición; en producción, de las variables del proveedor. Esta decisión, tomada antes de escribir la primera línea de lógica, es la que permitió contenerizar y desplegar sin modificar código de aplicación.

De la misma forma, el puerto de escucha se lee de `PORT`, lo que permite que los proveedores que asignan puerto dinámicamente funcionen sin configuración adicional.

### Sondas de salud y resolución de nombres

Los `HEALTHCHECK` apuntan a `127.0.0.1` y no a `localhost`.

El motivo surgió durante la puesta en marcha: dentro del contenedor, `localhost` resolvía a la dirección IPv6 `::1`, mientras que el listener de Fiber atendía únicamente en IPv4. La sonda fallaba aunque el servicio respondiera con normalidad a las peticiones reales. Fijar la dirección IPv4 explícita elimina la ambigüedad de resolución.

Las sondas también respetan la variable `PORT`, de modo que funcionan tanto en composición local como en plataformas que asignan el puerto en tiempo de ejecución.

---

## 8. Precisión numérica

Dos comportamientos esperados de la aritmética de punto flotante:

**Decimales largos.** Donde matemáticamente correspondería `14`, el resultado puede ser `14.000000000000002`. Es acumulación normal de error. No se redondea: alterar el valor ocultaría la precisión real del cálculo.

**Cero negativo.** Las reflexiones de Householder producen `-0`, que es matemáticamente cero pero se serializa distinto en JSON. Se normaliza en `fromDense`, en el punto donde los datos de gonum se convierten al tipo del dominio, aprovechando que IEEE 754 define `-0 + 0 = +0`.

---

## 9. Documentación relacionada

| Documento | Contenido |
|---|---|
| [`docs/requerimientos-generales.md`](docs/requerimientos-generales.md) | Requerimientos transversales, contrato entre APIs, trazabilidad al enunciado |
| [`docs/sustento-rotacion-qr.md`](docs/sustento-rotacion-qr.md) | Recorrido matemático paso a paso de la decisión de rotación condicional |
| [`fiber-api/ARCHITECTURE.md`](fiber-api/ARCHITECTURE.md) | Arquitectura interna del servicio de factorización |
| [`express-api/ARCHITECTURE.md`](express-api/ARCHITECTURE.md) | Arquitectura interna del servicio de estadísticas |
