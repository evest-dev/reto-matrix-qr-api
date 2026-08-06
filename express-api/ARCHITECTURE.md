# Arquitectura · express-api

Servicio en Node.js que recibe las matrices producidas por la factorización QR y calcula cinco estadísticas descriptivas sobre sus valores.

Las decisiones transversales al sistema están en [`../ARCHITECTURE.md`](../ARCHITECTURE.md).

---

## Responsabilidad y alcance

Este servicio **no conoce el origen de las matrices** ni la operación que las produjo. Recibe matrices, calcula números.

Esa ignorancia es deliberada: mantiene el servicio simple, testeable de forma aislada, y reutilizable si otro productor le enviara matrices distintas.

---

## Stack

| Componente | Versión | Motivo |
|---|---|---|
| Node | 20 | — |
| TypeScript | 5 en modo `strict` | Tipado explícito, coherente con el rigor de la API en Go |
| Express | 5 | — |
| Zod | 4 | Validación en el borde con tipos inferidos del esquema |
| Jest + Supertest | — | Pruebas unitarias y de integración |
| tsx | — | Ejecución en desarrollo sin paso de compilación |

Las versiones mayores de `express` y `@types/express` deben coincidir.

---

## Estructura y regla de dependencia

```
src/domain/
  matrix.ts                         entidad, verificación de diagonal, aplanado
  errors.ts                         jerarquía de errores con su código HTTP
  validation.ts                     coherencia matemática de las matrices
src/service/
  statistics.ts                     cálculo de las cinco métricas
src/routes/
  index.ts                          router de /api/v1
  statistics/
    index.ts                        definición de rutas
    controller.ts                   adaptador HTTP
    schema.ts                       esquema Zod
src/middleware/
  errorHandler.ts                   traducción de errores a HTTP
  validateBody.ts                   validación genérica con Zod
  requestLogger.ts                  registro de peticiones
src/config/index.ts                 lectura del entorno
src/app.ts                          composición, sin escuchar
src/server.ts                       arranque
```

Las dependencias apuntan hacia adentro:

```
routes ──► service ──► domain
   │           │
middleware ────┴──► (domain no importa Express ni Zod)
```

Espeja la separación de `fiber-api` —dominio, casos de uso, transporte— según el requerimiento de coherencia estructural entre servicios.

---

## Reglas de negocio

1. Las estadísticas se calculan sobre el **conjunto unificado** de valores de todas las matrices recibidas, no por matriz individual. El enunciado dice "todos los valores de las matrices", en plural.
2. `isAnyDiagonal` es un **único booleano**: verdadero si al menos una matriz recibida es diagonal.
3. Una matriz es diagonal cuando todo elemento fuera de la diagonal principal es cero. Los valores sobre la diagonal pueden ser cualquier número, incluido el cero: `[[5,0],[0,0]]` es diagonal.
4. El criterio se aplica igual a matrices no cuadradas.
5. Los valores son flotantes de doble precisión: la factorización QR produce decimales y negativos aunque la entrada original sea de enteros positivos.
6. El promedio no se redondea al calcularlo. El redondeo, si se aplicara, sería responsabilidad de la presentación.

---

## Endpoints

| Método | Ruta | Descripción |
|---|---|---|
| POST | `/api/v1/statistics` | Calcula las cinco métricas sobre las matrices recibidas |
| GET | `/health` | Estado del servicio, sin versionar |

### Entrada

```json
{
  "matrices": [
    { "name": "Q", "data": [[-0.6,-0.8,0],[-0.8,0.6,0],[0,0,1]] },
    { "name": "R", "data": [[-5,-4],[0,3],[0,0]] }
  ]
}
```

### Salida

```json
{
  "max": 3,
  "min": -5,
  "sum": -6.6,
  "average": -0.44,
  "isAnyDiagonal": false
}
```

**Sin envoltura.** `fiber-api` reenvía este objeto tal cual al cliente bajo la clave `statistics`. Envolverlo en `{ success, data }` obligaría a Go a desenvolverlo, añadiendo acoplamiento sin beneficio. Los errores sí llevan envoltura, homogénea con la API en Go.

---

## Decisiones de diseño

### Validación en dos capas

Zod valida la **forma del JSON** en el borde: que `matrices` sea un arreglo de objetos con `name` y `data`, y que `data` contenga números.

El dominio valida la **coherencia matemática**: que cada matriz sea rectangular y que sus valores sean finitos.

Son responsabilidades distintas y fallan con mensajes distintos. Una entrada como `[[1,2],[3]]` supera la validación de Zod —estructuralmente es un arreglo de arreglos de números— y es rechazada por el dominio, que identifica la matriz y la fila exactas.

### Tipos inferidos del esquema

El tipo de TypeScript se obtiene con `z.infer` a partir del esquema Zod, sin declararlo por separado. Así el tipo no puede divergir de la validación.

### Aplicación separada del servidor

`app.ts` construye y exporta la aplicación sin invocar `listen`; `server.ts` la importa y la levanta.

Esta separación permite a Supertest ejercitar los endpoints en las pruebas sin abrir un puerto real, evitando conflictos y esperas en la suite.

### Un solo recorrido

Máximo, mínimo y suma se obtienen en una única pasada sobre los valores, no en tres recorridos separados.

### `noUncheckedIndexedAccess` activado

Obliga a comprobar que un acceso por índice existe antes de usarlo. Es más verboso, pero elimina por construcción la clase de error más común al recorrer matrices.

Efecto secundario: genera guardas defensivas inalcanzables en tiempo de ejecución, que inflan artificialmente el denominador de la cobertura de ramas.

---

## Manejo de errores

Los errores de dominio son clases que extienden `AppError` y portan su código HTTP. La traducción a respuesta ocurre en un único punto: el middleware `errorHandler`, registrado al final de la cadena. Ninguna ruta decide códigos de estado por su cuenta.

```json
{ "success": false, "error": "mensaje en español" }
```

En producción no se exponen trazas de pila ni detalles internos.

---

## Ejecución

```bash
npm run dev          # desarrollo con recarga
npm run build        # compilar a dist/
npm start            # ejecutar lo compilado
npm test             # pruebas
npm test -- --coverage
npm run typecheck    # verificar tipos sin emitir
```

### Variables de entorno

| Variable | Valor por defecto | Uso |
|---|---|---|
| `PORT` | `4000` | Puerto de escucha |
| `NODE_ENV` | `development` | Controla el detalle de los errores expuestos |

---

## Convenciones

- Sin usos de `any`. Los datos externos entran como `unknown` y se estrechan mediante validación.
- Funciones puras en `service`: reciben datos, devuelven datos, sin acceder a `req` ni `res`.
- Comparación de flotantes en pruebas con `toBeCloseTo`, nunca `toBe`.
- Comentarios y mensajes de error en español; identificadores y campos JSON en inglés.

---

## Nota sobre las pruebas

Las pruebas unitarias reciben las matrices **escritas literalmente en el propio test**, no generadas por gonum. Prueban el cálculo estadístico de forma aislada, y por eso son independientes del algoritmo de factorización que use el servicio de Go.

Esto explica una diferencia que podría parecer inconsistencia: las pruebas usan los valores de Gram-Schmidt (`max 5`, `min -0.8`, `sum 14.2`) mientras que el flujo real produce los de Householder (`max 3`, `min -5`, `sum -6.6`). Ambos conjuntos son correctos para su respectiva factorización.

Detalle en [`../ARCHITECTURE.md`](../ARCHITECTURE.md#6-sobre-la-no-unicidad-de-la-factorización-qr).
