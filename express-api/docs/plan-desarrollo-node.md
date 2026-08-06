# Plan de desarrollo — Paso 2 · API en Node.js (Express + TypeScript)

Continuación de `../../fiber-api/docs/plan-desarrollo-go.md`. Cada etapa indica qué construir, qué pedirle a Claude Code, cómo verificarlo, y los conceptos que aparecen.

**Precondición:** el paso 1 debe estar completo. `fiber-api` levanta y responde en `/api/v1/matrix/rotate`; su endpoint `/process` falla al llamar a este servicio, y eso es justamente lo que vas a resolver aquí.

**Principio rector:** igual que en Go, de adentro hacia afuera. Dominio primero, transporte al final.

---

## Etapa 0 · Preparar el proyecto

### 0.1 Limpiar lo que existe

El `index.js` de prueba se elimina: la estructura definitiva vive en `src/`.

```bash
cd express-api
rm -f index.js
```

### 0.2 Crear la estructura

```bash
mkdir -p src/domain src/service src/config src/middleware src/routes/statistics
mkdir -p tests/unit tests/integration
```

### 0.3 Dependencias

```bash
npm install express cors helmet zod
npm install -D typescript tsx @types/node @types/express @types/cors \
  jest ts-jest @types/jest supertest @types/supertest
```

**Qué es cada una.** `express` es el framework; `cors` y `helmet` son middleware de seguridad; `zod` valida la entrada. En desarrollo: `typescript` y `tsx` para compilar y ejecutar, `jest` con `ts-jest` para pruebas, `supertest` para probar endpoints, y los paquetes `@types/*` que aportan las definiciones de tipos de librerías escritas en JavaScript.

### 0.4 Configuración de TypeScript, Jest y scripts

**Prompt:**

```
Lee express-api/CLAUDE.md y express-api/docs/diseno-tecnico.md.

Crea en express-api los archivos tsconfig.json, jest.config.js y actualiza
package.json con los scripts, según la sección 5 del diseño técnico.

Solo esos tres archivos de configuración. No implementes código todavía.
```

**Verificación:**

```bash
npx tsc --noEmit
```

Sin errores. Aún no hay código que compilar, pero confirma que la configuración es válida.

---

## Conceptos que vas a encontrar

Breve, para reconocerlos cuando aparezcan.

### El modo estricto y `noUncheckedIndexedAccess`

Con esta opción activada, acceder a `matriz[0]` no devuelve `number[]` sino `number[] | undefined`. TypeScript te obliga a comprobar que existe antes de usarlo:

```typescript
const row = matrix[i];
if (row === undefined) continue;
```

Es más verboso, y al principio molesta. La razón de tenerlo activo: elimina por construcción el error más frecuente al recorrer matrices.

### Tipos inferidos desde Zod

En vez de declarar el tipo y el validador por separado, se declara el esquema y el tipo se deduce:

```typescript
const schema = z.object({ name: z.string() });
type Request = z.infer<typeof schema>;
```

Así el tipo no puede quedar desincronizado de la validación.

### `readonly` en los tipos

`readonly number[]` indica que el arreglo no se modificará. En el dominio se usa para dejar explícito que los datos de entrada no se mutan: las funciones reciben y devuelven, no alteran.

### El middleware de errores lleva cuatro parámetros

Express distingue un manejador de errores de un middleware normal **por la cantidad de parámetros**: si recibe cuatro `(error, req, res, next)`, es de errores. Por eso el cuarto parámetro se declara aunque no se use.

### `next(error)` en lugar de responder

Cuando algo falla, el controlador no construye la respuesta de error: llama a `next(error)` y deja que el manejador central decida. Así el formato de error es uno solo en toda la aplicación.

---

## Etapa 1 · El dominio

**Qué construyes:** el tipo `Matrix`, la verificación de diagonal, el aplanado y la jerarquía de errores. Sin Express, sin Zod.

**Archivos:** `src/domain/matrix.ts`, `src/domain/errors.ts`, `src/domain/validation.ts`

**Prompt:**

```
Lee express-api/CLAUDE.md, express-api/docs/requerimientos-paso-2.md y
express-api/docs/diseno-tecnico.md.

Implementa la etapa 1: src/domain/matrix.ts, src/domain/errors.ts y
src/domain/validation.ts, según las secciones 6.1, 6.2 y 6.3 del diseño técnico.

Solo esos tres archivos. El dominio no debe importar Express ni Zod.
```

**Qué revisar:**

- ¿`isDiagonal` devuelve `true` para `[[5,0],[0,0]]`? Un cero en la diagonal no la invalida.
- ¿`isDiagonal` devuelve `true` para `[[7]]`? No tiene elementos fuera de la diagonal.
- ¿`flattenAll` preserva el orden y no pierde elementos?
- ¿Los errores extienden `AppError` y llevan su `statusCode`?
- ¿Ningún import de `express` ni `zod`?

**Verificación:**

```bash
npm run typecheck
```

---

## Etapa 2 · Las pruebas del dominio

**Qué construyes:** las pruebas antes de seguir. La lógica de diagonal tiene casos de borde que conviene fijar ahora.

**Archivo:** `tests/unit/matrix.test.ts`

**Prompt:**

```
Implementa tests/unit/matrix.test.ts según la sección 7.1 del diseño técnico.
Incluye los cinco casos de referencia de matriz diagonal del requerimiento RF-06.
```

**Verificación:**

```bash
npm test
```

Deberías ver los casos con su nombre y el resultado de cada uno. Si `isDiagonal` falla en algún borde, se arregla aquí y no más adelante.

---

## Etapa 3 · El servicio de estadísticas

**Qué construyes:** el cálculo de las cinco métricas. Función pura: recibe matrices, devuelve números.

**Archivos:** `src/service/statistics.ts`, `tests/unit/statistics.test.ts`

**Prompt:**

```
Implementa src/service/statistics.ts y tests/unit/statistics.test.ts según
las secciones 6.4 y 7.2 del diseño técnico.

Máximo, mínimo y suma deben calcularse en un único recorrido, no en tres.
La prueba del conjunto de referencia debe verificar los cinco valores esperados.
```

**El detalle que importa:** la prueba con el conjunto de referencia. Son las matrices Q y R que produce `fiber-api` con la matriz `[[3,0],[4,5],[0,0]]`, y sus valores esperados son:

| Métrica | Valor |
|---|---|
| max | 5 |
| min | −0.8 |
| sum | 14.2 |
| average | 0.9466666666666667 |
| isAnyDiagonal | false |

Si esa prueba pasa, tu servicio produce exactamente lo que la API en Go espera recibir.

**Verificación:**

```bash
npm test
```

Ojo con las comparaciones: `toBeCloseTo`, nunca `toBe`. Con flotantes, `0.1 + 0.2` no es exactamente `0.3` en ningún lenguaje.

---

## Etapa 4 · La capa HTTP

**Qué construyes:** esquema Zod, middleware, controlador, rutas, configuración y la aplicación.

**Archivos:** `src/routes/statistics/{schema,controller,index}.ts`, `src/routes/index.ts`, `src/middleware/{validateBody,errorHandler,requestLogger}.ts`, `src/config/index.ts`, `src/app.ts`, `src/server.ts`

**Prompt:**

```
Implementa la capa HTTP según las secciones 6.5 a 6.13 del diseño técnico:
esquema Zod, los tres middleware, el controlador, las rutas, la configuración,
app.ts y server.ts.

app.ts debe exportar la aplicación sin llamar a listen; server.ts la levanta.
La respuesta exitosa es el objeto de estadísticas sin envolver.
```

**Los dos detalles a entender:**

**Por qué `app.ts` no llama a `listen`.** Porque Supertest necesita la aplicación, no un servidor escuchando. Si `app.ts` abriera el puerto, cada prueba tendría que gestionarlo y la suite se volvería lenta y frágil.

**Por qué la respuesta va sin envoltura.** La API en Go reenvía este objeto tal cual bajo la clave `statistics`. Si lo envolvieras en `{ success, data }`, Go tendría que desenvolverlo. Los errores sí van envueltos, en el mismo formato que la API en Go.

**Verificación:**

```bash
npm run dev
```

En otra terminal:

```bash
curl http://localhost:4000/health
```

Debe responder `{"status":"ok"}`.

Y el cálculo real:

```bash
curl -X POST http://localhost:4000/api/v1/statistics \
  -H "Content-Type: application/json" \
  -d '{"matrices":[{"name":"Q","data":[[0.6,-0.8,0],[0.8,0.6,0],[0,0,1]]},{"name":"R","data":[[5,4],[0,3],[0,0]]}]}'
```

Debe devolver máximo 5, mínimo −0.8, suma 14.2 y `isAnyDiagonal` falso.

Prueba también que rechace bien:

```bash
curl -X POST http://localhost:4000/api/v1/statistics \
  -H "Content-Type: application/json" \
  -d '{"matrices":[{"name":"Q","data":[[1,2],[3]]}]}'
```

Debe responder 400 mencionando que no es rectangular.

---

## Etapa 5 · Pruebas de integración

**Qué construyes:** las pruebas del endpoint completo con Supertest.

**Archivo:** `tests/integration/statistics.route.test.ts`

**Prompt:**

```
Implementa tests/integration/statistics.route.test.ts según la sección 7.3
del diseño técnico. Incluye la prueba que verifica que la respuesta tiene
exactamente las cinco claves, sin envoltura.
```

**Verificación:**

```bash
npm test -- --coverage
```

Toda la suite en verde y la cobertura por encima del umbral configurado.

---

## Etapa 6 · Conectar ambas APIs

Aquí es donde el paso 1 y el paso 2 se encuentran.

**Levanta ambos servicios**, cada uno en su terminal:

```bash
# terminal 1
cd fiber-api && go run ./cmd/api

# terminal 2
cd express-api && npm run dev
```

**Prueba el flujo completo** contra la API en Go:

```bash
curl -X POST http://localhost:3000/api/v1/matrix/process \
  -H "Content-Type: application/json" \
  -d '{"matrix": [[3,0],[4,5],[0,0]]}'
```

Debe devolver, en una sola respuesta: la matriz original, `wasRotated` en falso, `factorizedFrom` como `"original"`, las matrices Q y R, y las cinco estadísticas.

**Prueba el camino de la rotación**, con una matriz ancha:

```bash
curl -X POST http://localhost:3000/api/v1/matrix/process \
  -H "Content-Type: application/json" \
  -d '{"matrix": [[0,5,0],[3,4,0]]}'
```

Debe devolver `wasRotated` en verdadero, `factorizedFrom` como `"rotated"`, el campo `rotated` con `[[3,0],[4,5],[0,0]]`, y las mismas estadísticas del caso anterior — porque tras rotar es la misma matriz.

**Si esas dos peticiones responden bien, el desarrollo está completo.** Lo que sigue es empaquetado y despliegue.

---

## Etapa 7 · Cierre

En orden:

1. **Docker** — un `Dockerfile` por servicio y el `docker-compose.yml` en la raíz. Como la URL sale de variable de entorno, no hay que tocar código.
2. **README** — instalación, ejecución, ejemplos de petición y respuesta, y las decisiones de interpretación del enunciado.
3. **Despliegue en la nube** — requisito obligatorio.
4. **JWT** — opcional.
5. **Frontend** — opcional.

---

## Comandos de referencia

```bash
npm run dev             # desarrollo con recarga
npm run build           # compilar a dist/
npm start               # ejecutar lo compilado
npm test                # pruebas
npm test -- --coverage  # con cobertura
npm run test:watch      # en modo watch
npm run typecheck       # verificar tipos sin emitir
```

---

## Errores frecuentes con TypeScript estricto

**"Object is possibly 'undefined'"** — es `noUncheckedIndexedAccess` haciendo su trabajo. Accediste a un índice sin comprobar que existe. Guarda el valor en una variable y verifica antes de usarlo.

**"Property does not exist on type"** — falta el paquete `@types/` de esa librería, o el objeto no tiene ese campo. Revisa que instalaste los tipos.

**Jest no encuentra los tests** — deben estar bajo `tests/`, según el `roots` del `jest.config.js`, y terminar en `.test.ts`.

**"Cannot find module"** — los imports relativos necesitan la ruta correcta desde el archivo actual. Desde `src/routes/statistics/controller.ts` hacia el servicio son dos niveles: `../../service/statistics`.

**Las comparaciones de flotantes fallan** — usa `toBeCloseTo(esperado, 9)` en lugar de `toBe`. Es aritmética de punto flotante, no un error tuyo.

---

## Regla de trabajo

Igual que en el paso 1: después de cada etapa, lee el código, córrelo, y si algo no entiendes, pregunta antes de avanzar. Vas a tener que defender estas decisiones en la entrevista.
