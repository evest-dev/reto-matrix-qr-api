# Requerimientos — Paso 2 · API en Node.js (Express)

Reto técnico Interseguro · División TI

Requerimientos específicos de la API en Node.js. Los transversales a ambos servicios están en `../../docs/requerimientos-generales.md`.

La columna **fase** indica cuándo se aborda cada requerimiento: *local* durante el desarrollo, *cierre* una vez que ambas APIs funcionan en la máquina.

---

## Continuidad con el paso 1

Este documento continúa donde termina `../../fiber-api/docs/requerimientos-paso-1.md`.

**Precondición para empezar.** El paso 1 debe estar completo y verificado: `fiber-api` levanta, valida matrices, rota condicionalmente y produce Q y R. Su endpoint `POST /api/v1/matrix/process` falla al invocar a este servicio — ese fallo es exactamente lo que el paso 2 resuelve.

**Punto de empalme.** El paso 1 termina produciendo el cuerpo que este servicio recibe:

```json
{ "matrices": [ { "name": "Q", "data": [[...]] }, { "name": "R", "data": [[...]] } ] }
```

Definido en `RF-05 · Consumo de la API de estadísticas` del paso 1, y en el contrato de `../../docs/requerimientos-generales.md`.

**Criterio de cierre del paso 2.** El flujo completo responde de extremo a extremo:

```bash
curl -X POST http://localhost:3000/api/v1/matrix/process \
  -H "Content-Type: application/json" \
  -d '{"matrix": [[3,0],[4,5],[0,0]]}'
```

Debe devolver original, Q, R y las cinco estadísticas en una sola respuesta.

Completado el paso 2, sigue la fase de cierre: contenerización, README, despliegue en nube y opcionales.

---

## Responsabilidad del servicio

Recibir las matrices producidas por la factorización QR de `fiber-api`, calcular cinco estadísticas descriptivas sobre sus valores, y devolverlas.

Este servicio no conoce el origen de las matrices ni la operación que las produjo. Recibe matrices, calcula números. Esa ignorancia deliberada lo mantiene simple y testeable.

---

## RF-01 · Recepción y validación de las matrices

**Origen:** pág. 3 — "reciba el resultado de las matrices devueltas por la primera API"
**Fase:** local
**Etapa del plan:** 1 y 2

Acepta una o más matrices identificadas por nombre.

**Criterios de aceptación**

- Acepta un arreglo de objetos con `name` (texto) y `data` (arreglo de arreglos de números).
- Acepta matrices de cualquier dimensión, incluidas las de un solo elemento.
- Acepta valores decimales y negativos.
- Rechaza con HTTP 400 un cuerpo que no sea JSON válido.
- Rechaza con HTTP 400 un arreglo `matrices` vacío o ausente.
- Rechaza con HTTP 400 una matriz vacía, o con filas de longitudes distintas.
- Rechaza con HTTP 400 valores no numéricos, incluidos `NaN` e `Infinity`.
- El mensaje de error indica qué matriz y qué fila causaron el rechazo.
- La validación de forma ocurre en el borde, antes de llegar a la lógica de negocio.

**Verificación**

```bash
curl -X POST http://localhost:4000/api/v1/statistics \
  -H "Content-Type: application/json" \
  -d '{"matrices": [{"name":"Q","data":[[1,2],[3]]}]}'
# 400 · indica que la fila 1 de la matriz Q tiene 1 columna y se esperaban 2
```

---

## RF-02 · Valor máximo

**Origen:** pág. 4 — "Valor máximo: El valor máximo encontrado en las matrices"
**Fase:** local
**Etapa del plan:** 3

**Criterios de aceptación**

- Devuelve el mayor valor del conjunto unificado de todas las matrices.
- Funciona con valores negativos: el máximo de `[-5, -2]` es `-2`.
- Con un único valor, ese valor es el máximo.

---

## RF-03 · Valor mínimo

**Origen:** pág. 4 — "Valor mínimo: El valor mínimo encontrado en las matrices"
**Fase:** local
**Etapa del plan:** 3

**Criterios de aceptación**

- Devuelve el menor valor del conjunto unificado.
- Funciona con valores negativos y decimales.

---

## RF-04 · Suma total

**Origen:** pág. 4 — "Suma total: La suma total de todos los valores de las matrices"
**Fase:** local
**Etapa del plan:** 3

**Criterios de aceptación**

- Suma todos los valores de todas las matrices.
- Los valores negativos restan.
- Para el conjunto de referencia (Q y R del caso 3×2), la suma es `14.2`.

---

## RF-05 · Promedio

**Origen:** pág. 4 — "Promedio: El promedio de todos los valores de las matrices"
**Fase:** local
**Etapa del plan:** 3

**Criterios de aceptación**

- Media aritmética: suma dividida entre la cantidad total de valores.
- La cantidad considera todos los elementos de todas las matrices, incluidos los ceros.
- No se redondea al calcular. Para el conjunto de referencia: `14.2 / 15 = 0.9466666666666667`.

---

## RF-06 · Verificación de matriz diagonal

**Origen:** pág. 4 — "Matriz diagonal: Verificar si alguna matriz es diagonal"
**Fase:** local
**Etapa del plan:** 3

**Criterios de aceptación**

- Una matriz es diagonal cuando todo elemento fuera de la diagonal principal es cero.
- La diagonal principal son los elementos en posición `[i][i]`.
- Los valores sobre la diagonal pueden ser cualquier número, incluido el cero.
- Devuelve un único booleano: verdadero si al menos una matriz del conjunto es diagonal.
- El criterio se aplica igual a matrices no cuadradas.

**Casos de referencia**

| Matriz | ¿Diagonal? | Motivo |
|---|---|---|
| `[[5,0],[0,3]]` | sí | fuera de la diagonal solo hay ceros |
| `[[5,0],[0,0]]` | sí | un cero en la diagonal no la invalida |
| `[[5,7],[0,3]]` | no | el 7 está fuera de la diagonal |
| `[[5,4],[0,3],[0,0]]` | no | el 4 está fuera de la diagonal |
| `[[7]]` | sí | no tiene elementos fuera de la diagonal |

---

## RF-07 · Composición de la respuesta

**Origen:** pág. 2 — "devolverá estas estadísticas como resultado"
**Fase:** local
**Etapa del plan:** 4

**Criterios de aceptación**

- Devuelve exactamente cinco campos: `max`, `min`, `sum`, `average`, `isAnyDiagonal`.
- El objeto no se envuelve en `{ success, data }`: `fiber-api` lo reenvía tal cual bajo la clave `statistics`, y envolverlo rompería el contrato.
- Campos en `camelCase`.
- Responde HTTP 200 ante cálculo exitoso.

**Forma de la respuesta**

```json
{
  "max": 5,
  "min": -0.8,
  "sum": 14.2,
  "average": 0.9466666666666667,
  "isAnyDiagonal": false
}
```

---

## RNF-01 · Arquitectura por capas

**Origen:** pág. 4 — "se espera coherencia en su estructura"
**Fase:** local

**Criterios de aceptación**

- Separación en `domain`, `service`, `routes`, `middleware` y `config`, espejando la estructura de `fiber-api`.
- Las dependencias apuntan hacia adentro: `domain` y `service` no importan Express ni Zod.
- Las funciones de `service` son puras: reciben datos y devuelven datos, sin acceder a `req` ni `res`.
- La aplicación se exporta sin escuchar (`app.ts`) y se levanta aparte (`server.ts`), para permitir pruebas sin abrir puerto.

---

## RNF-02 · Tipado estricto

**Origen:** pág. 2 — "siguiendo las mejores prácticas de codificación"
**Fase:** local

**Criterios de aceptación**

- TypeScript con `strict: true`.
- Sin usos de `any`. Los datos externos entran como `unknown` y se estrechan mediante validación.
- Los tipos de los datos validados se infieren del esquema Zod, sin declararlos por duplicado.
- `npm run typecheck` termina sin errores.

---

## RNF-03 · Manejo de errores

**Origen:** pág. 2 — "siguiendo las mejores prácticas de codificación"
**Fase:** local

**Criterios de aceptación**

- Errores de dominio como clases que extienden una base común y portan su código HTTP.
- La traducción a respuesta HTTP ocurre en un único middleware, registrado al final de la cadena.
- Formato uniforme y homogéneo con `fiber-api`: `{ "success": false, "error": "mensaje" }`.
- En producción no se exponen trazas de pila ni detalles internos.
- Los errores asíncronos se propagan a `next()`, no se tragan.

---

## RNF-04 · Documentación del código

**Origen:** pág. 2 — "Documentar el código de manera clara y concisa"
**Fase:** local

**Criterios de aceptación**

- Comentarios JSDoc en las funciones exportadas de `domain` y `service`.
- Los comentarios explican el porqué de las decisiones, no lo que el código ya dice.
- La función que verifica si una matriz es diagonal documenta el criterio aplicado.

---

## RNF-05 · Pruebas

**Origen:** pág. 3 — funcionalidad opcional
**Fase:** local

**Criterios de aceptación**

- Pruebas unitarias de cada estadística con casos de borde: valores negativos, decimales, un único elemento, ceros.
- La verificación de diagonal se prueba con los cinco casos de referencia de RF-06.
- Prueba de integración del endpoint con Supertest, sobre la aplicación exportada.
- Prueba que confirma el rechazo de entradas malformadas con HTTP 400.
- Comparación de flotantes con `toBeCloseTo`, nunca `toBe`.
- Existe una prueba con el conjunto de referencia completo (Q y R del caso 3×2) que verifica los cinco valores esperados de una vez.
- Suite ejecutable con `npm test`.

---

## RNF-06 · Contenerización

**Origen:** pág. 2 — "Utilizar Docker para contenerizar las aplicaciones"
**Fase:** cierre

**Criterios de aceptación**

- `Dockerfile` con build multietapa: compilación de TypeScript en una etapa, ejecución de `dist/` en otra.
- La imagen final no incluye dependencias de desarrollo ni código fuente TypeScript.
- Usuario sin privilegios.
- Healthcheck apuntando a `GET /health`.

---

## Trazabilidad

| Requerimiento | Origen | Fase | Etapa |
|---|---|---|---|
| RF-01 Recepción y validación | pág. 3 | local | 1–2 |
| RF-02 Valor máximo | pág. 4 | local | 3 |
| RF-03 Valor mínimo | pág. 4 | local | 3 |
| RF-04 Suma total | pág. 4 | local | 3 |
| RF-05 Promedio | pág. 4 | local | 3 |
| RF-06 Matriz diagonal | pág. 4 | local | 3 |
| RF-07 Composición de respuesta | pág. 2 | local | 4 |
| RNF-01 Arquitectura por capas | pág. 4 | local | 1–5 |
| RNF-02 Tipado estricto | pág. 2 | local | 1–5 |
| RNF-03 Manejo de errores | pág. 2 | local | 4 |
| RNF-04 Documentación | pág. 2 | local | 1–5 |
| RNF-05 Pruebas | pág. 3 | local | 3, 5 |
| RNF-06 Contenerización | pág. 2 | cierre | — |

---

## Conjuntos de referencia para verificación

Existen dos conjuntos válidos, según el algoritmo que produzca la factorización. Ambos satisfacen `Q × R = A`. Las pruebas unitarias usan el primero por su legibilidad; el flujo de extremo a extremo produce el segundo.

### A · Gram-Schmidt (usado en las pruebas unitarias)

Es el resultado del cálculo manual, con signos positivos en la primera columna. Se emplea en los tests porque los números son más fáciles de seguir y verificar a mano.

```
Q = 0.6  -0.8  0        R = 5  4
    0.8   0.6  0            0  3
    0     0    1            0  0
```

Los 15 valores del conjunto unificado:

```
0.6  -0.8  0  0.8  0.6  0  0  0  1  |  5  4  0  3  0  0
```

| Métrica | Valor |
|---|---|
| `max` | 5 |
| `min` | −0.8 |
| `sum` | 14.2 |
| `average` | 0.9466666666666667 |
| `isAnyDiagonal` | false |

### B · Householder (salida real de gonum, flujo de extremo a extremo)

Es lo que produce `fiber-api` al procesar `[[3,0],[4,5],[0,0]]`. Gonum emplea reflexiones de Householder, numéricamente más estables, que invierten el signo de la primera columna de Q y de la primera fila de R.

```
Q = -0.6  -0.8  0       R = -5  -4
    -0.8   0.6  0            0   3
     0     0    1            0   0
```

Los 15 valores del conjunto unificado:

```
-0.6  -0.8  0  -0.8  0.6  0  0  0  1  |  -5  -4  0  3  0  0
```

| Métrica | Valor |
|---|---|
| `max` | 3 |
| `min` | −5 |
| `sum` | −6.6 |
| `average` | −0.44 |
| `isAnyDiagonal` | false |

### Por qué ambos son correctos

La factorización QR no es única: está determinada salvo el signo de cada columna de Q, compensado por el signo de la fila correspondiente de R. Distintos algoritmos eligen distintos signos y todos satisfacen la identidad fundamental.

Comprobación con el conjunto B:

```
(−0.6)(−5) + (−0.8)(0) + (0)(0) = 3
(−0.6)(−4) + (−0.8)(3) + (0)(0) = 2.4 − 2.4 = 0
(−0.8)(−5) + ( 0.6)(0) + (0)(0) = 4
(−0.8)(−4) + ( 0.6)(3) + (0)(0) = 3.2 + 1.8 = 5
```

Se recupera exactamente `[[3,0],[4,5],[0,0]]`.

**Consecuencia para las pruebas de este servicio.** Los tests unitarios reciben las matrices escritas literalmente en el propio test, no generadas por gonum: prueban el cálculo estadístico de forma aislada. Por eso usan el conjunto A y siguen siendo válidos aunque el flujo real produzca el B.

En ambos conjuntos `isAnyDiagonal` es falso: Q tiene valores fuera de la diagonal y R tiene un valor fuera de la diagonal en la primera fila.

---

## Endpoints

| Método | Ruta | Descripción |
|---|---|---|
| POST | `/api/v1/statistics` | Calcula las cinco métricas sobre las matrices recibidas |
| GET | `/health` | Estado del servicio, sin versionar |
