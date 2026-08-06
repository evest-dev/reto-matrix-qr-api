# API en Node.js (Express) — Diseño técnico

Reto técnico Interseguro · División TI

---

## 1. Responsabilidad y alcance

Este servicio recibe matrices y calcula cinco estadísticas descriptivas sobre sus valores. No conoce el origen de las matrices ni la operación que las produjo.

Esa ignorancia es deliberada: mantiene el servicio simple, testeable de forma aislada, y reutilizable si mañana otro productor le enviara matrices distintas.

---

## 2. Decisiones de diseño

### 2.1 Conjunto unificado, no por matriz

El enunciado dice "el promedio de todos los valores de **las matrices**", en plural. Las cinco métricas se calculan sobre un único conjunto que reúne los valores de todas las matrices recibidas.

Consecuencia práctica: aplanar primero, calcular después. Un solo recorrido produce máximo, mínimo y suma simultáneamente.

### 2.2 Un booleano para la diagonal

El enunciado dice "verificar si **alguna** matriz es diagonal". Se devuelve un único booleano para el conjunto, no un valor por matriz.

### 2.3 Respuesta sin envoltura

`fiber-api` reenvía este objeto tal cual al cliente bajo la clave `statistics`. Envolverlo en `{ success, data }` obligaría a Go a desenvolverlo, añadiendo acoplamiento sin beneficio. La respuesta exitosa es el objeto de estadísticas directo.

Los errores sí llevan envoltura, homogénea con `fiber-api`: `{ success: false, error }`.

### 2.4 Aplicación separada del servidor

`app.ts` construye y exporta la aplicación Express sin llamar a `listen`. `server.ts` la importa y la levanta.

Esta separación es lo que permite a Supertest probar los endpoints sin abrir un puerto real, evitando conflictos y esperas en la suite de pruebas.

### 2.5 Validación en el borde con Zod

Los datos externos entran como `unknown` y se estrechan mediante un esquema Zod. El tipo de TypeScript se **infiere** del esquema con `z.infer`, no se declara por separado: así el esquema y el tipo no pueden divergir.

---

## 3. Flujo de ejecución

```
              ┌──────────────────────────────┐
              │  POST /api/v1/statistics     │
              │  { matrices: [ {name,data} ] }│
              └──────────────┬───────────────┘
                             │
              ┌──────────────▼───────────────┐
              │  validateBody (Zod)          │
              │  · forma del cuerpo          │
              │  · tipos numéricos           │
              └──────────────┬───────────────┘
                             │
              ┌──────────────▼───────────────┐
              │  domain · assertRectangular  │
              │  · filas homogéneas          │
              │  · sin NaN ni Infinity       │
              └──────────────┬───────────────┘
                             │
              ┌──────────────▼───────────────┐
              │  service · flatten           │
              │  Q(9) + R(6) → 15 valores    │
              └──────────────┬───────────────┘
                             │
              ┌──────────────▼───────────────┐
              │  service · calculate         │
              │  1 recorrido → max, min, sum │
              │  average = sum / n           │
              │  isAnyDiagonal = some(...)   │
              └──────────────┬───────────────┘
                             │
              ┌──────────────▼───────────────┐
              │  200 · { max, min, sum,      │
              │          average,            │
              │          isAnyDiagonal }     │
              └──────────────────────────────┘

  cualquier error ──► errorHandler ──► { success:false, error }
```

---

## 4. Estructura del proyecto

```
express-api/
├── src/
│   ├── domain/
│   │   ├── matrix.ts               entidad e invariantes
│   │   └── errors.ts               jerarquía de errores
│   ├── service/
│   │   └── statistics.ts           cálculo de las cinco métricas
│   ├── routes/
│   │   ├── index.ts                montaje del router v1
│   │   └── statistics/
│   │       ├── index.ts            definición de rutas
│   │       ├── controller.ts       adaptador HTTP
│   │       └── schema.ts           esquema Zod
│   ├── middleware/
│   │   ├── errorHandler.ts         traducción de errores a HTTP
│   │   ├── validateBody.ts         validación genérica con Zod
│   │   └── requestLogger.ts        registro de peticiones
│   ├── config/
│   │   └── index.ts                variables de entorno
│   ├── app.ts                      composición, sin listen
│   └── server.ts                   arranque
├── tests/
│   ├── unit/
│   │   ├── matrix.test.ts
│   │   └── statistics.test.ts
│   └── integration/
│       └── statistics.route.test.ts
├── Dockerfile
├── jest.config.js
├── tsconfig.json
└── package.json
```

**Regla de dependencia:** `domain` no importa nada. `service` importa `domain`. `routes` y `middleware` importan `service` y `domain`. Express y Zod solo aparecen de `routes` y `middleware` hacia afuera.

---

## 5. Configuración base

### 5.1 `package.json`

```json
{
  "name": "express-api",
  "version": "1.0.0",
  "description": "API de estadísticas sobre matrices",
  "main": "dist/server.js",
  "scripts": {
    "dev": "tsx watch src/server.ts",
    "build": "tsc",
    "start": "node dist/server.js",
    "test": "jest",
    "test:watch": "jest --watch",
    "typecheck": "tsc --noEmit"
  },
  "dependencies": {
    "cors": "^2.8.5",
    "express": "^5.0.0",
    "helmet": "^7.1.0",
    "zod": "^4.0.0"
  },
  "devDependencies": {
    "@types/cors": "^2.8.17",
    "@types/express": "^5.0.0",
    "@types/jest": "^29.5.12",
    "@types/node": "^20.14.0",
    "@types/supertest": "^6.0.2",
    "jest": "^29.7.0",
    "supertest": "^7.0.0",
    "ts-jest": "^29.1.5",
    "tsx": "^4.16.0",
    "typescript": "^5.5.0"
  }
}
```

**Express 5 y sus cambios respecto a la versión 4.** La versión mayor de `@types/express` debe coincidir con la de `express`: mezclar tipos de la v4 con la v5 produce errores de compilación difíciles de interpretar.

Ninguno de los cambios incompatibles de la v5 afecta este diseño:

| Cambio | Situación en este proyecto |
|---|---|
| `path-to-regexp` v8 exige nombre en los comodines (`/*splat`) | No se usan comodines: todas las rutas son cadenas literales |
| `req.body` es `undefined` si no hay parser, en vez de `{}` | `express.json()` está registrado; y si llegara `undefined`, Zod lo rechaza con 400, que es el comportamiento deseado |
| `app.del()` eliminado en favor de `app.delete()` | No se usa |
| `res.status()` valida el código recibido | Solo se emiten códigos válidos |
| Los errores asíncronos se propagan automáticamente al manejador | Vuelve redundante el `try/catch` del controlador, que se conserva por ser explícito y por no depender del comportamiento del framework |

**Sobre las versiones de Jest y Zod.** Jest 30 requiere ts-jest ≥ 29.4: la serie 29.4.x lo soporta, no hace falta degradar ninguno de los dos. Zod 4 renombró `ZodError.errors` a `ZodError.issues`; el código de este diseño usa `issues`, que existe en ambas versiones mayores.

### 5.2 `tsconfig.json`

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "commonjs",
    "lib": ["ES2022"],
    "types": ["node", "jest"],
    "rootDir": "./src",
    "outDir": "./dist",
    "strict": true,
    "noUncheckedIndexedAccess": true,
    "noImplicitOverride": true,
    "esModuleInterop": true,
    "forceConsistentCasingInFileNames": true,
    "skipLibCheck": true,
    "declaration": false,
    "sourceMap": true
  },
  "include": ["src/**/*"],
  "exclude": ["node_modules", "dist", "tests"]
}
```

`noUncheckedIndexedAccess` obliga a comprobar que un acceso por índice existe antes de usarlo. Es incómodo al principio, pero previene exactamente la clase de error que aparece al recorrer matrices.

**Sobre el campo `types`.** Declararlo desactiva la inclusión automática de todos los paquetes bajo `node_modules/@types` y limita el conjunto a los listados. Por eso se declaran ambos: `node` porque `domain/errors.ts` usa `Error.captureStackTrace`, y `jest` porque los archivos de prueba dependen de los globales `describe`, `it` y `expect`. Listar solo `node` haría fallar el chequeo de tipos en `tests/`.

### 5.3 `jest.config.js`

```js
/** @type {import('jest').Config} */
module.exports = {
  preset: 'ts-jest',
  testEnvironment: 'node',
  roots: ['<rootDir>/tests'],
  collectCoverageFrom: ['src/**/*.ts', '!src/server.ts'],
  coverageThreshold: {
    global: { branches: 80, functions: 80, lines: 80, statements: 80 },
  },
};
```

Se excluye `server.ts` de la cobertura: solo llama a `listen`, no tiene lógica que probar.

---

## 6. Código

### 6.1 Dominio — `src/domain/matrix.ts`

```typescript
/**
 * Entidades y reglas del dominio de matrices.
 * No depende de Express, Zod ni de ningún detalle de transporte.
 */

/** Una matriz es una lista de filas, cada una con la misma cantidad de números. */
export type Matrix = readonly (readonly number[])[];

/** Matriz identificada por nombre, tal como llega desde la API de factorización. */
export interface NamedMatrix {
  readonly name: string;
  readonly data: Matrix;
}

/** Cantidad de filas. */
export function rows(matrix: Matrix): number {
  return matrix.length;
}

/** Cantidad de columnas. Devuelve 0 para una matriz vacía. */
export function cols(matrix: Matrix): number {
  return matrix[0]?.length ?? 0;
}

/**
 * Determina si una matriz es diagonal.
 *
 * Una matriz es diagonal cuando todo elemento fuera de la diagonal principal
 * es cero. Los elementos sobre la diagonal pueden tomar cualquier valor,
 * incluido el cero: `[[5,0],[0,0]]` sigue siendo diagonal.
 *
 * El criterio se aplica igual a matrices no cuadradas. En la práctica, una
 * matriz no cuadrada con valores fuera de la diagonal no califica.
 */
export function isDiagonal(matrix: Matrix): boolean {
  for (let i = 0; i < matrix.length; i++) {
    const row = matrix[i];
    if (row === undefined) continue;

    for (let j = 0; j < row.length; j++) {
      if (i === j) continue; // la diagonal admite cualquier valor

      const value = row[j];
      if (value !== undefined && value !== 0) {
        return false;
      }
    }
  }

  return true;
}

/**
 * Aplana todas las matrices en un único arreglo de valores.
 *
 * El enunciado pide las estadísticas sobre "todos los valores de las
 * matrices", en plural: se tratan como un solo conjunto.
 */
export function flattenAll(matrices: readonly NamedMatrix[]): number[] {
  const values: number[] = [];

  for (const { data } of matrices) {
    for (const row of data) {
      values.push(...row);
    }
  }

  return values;
}
```

### 6.2 Errores — `src/domain/errors.ts`

```typescript
/**
 * Jerarquía de errores del dominio.
 * Cada error porta su código HTTP para que el middleware de errores
 * no tenga que decidirlo mediante condicionales.
 */

/** Error base. Todo error esperado del sistema desciende de esta clase. */
export abstract class AppError extends Error {
  abstract readonly statusCode: number;

  constructor(message: string) {
    super(message);
    this.name = new.target.name;
    Error.captureStackTrace(this, new.target);
  }
}

/** La entrada no cumple las reglas de forma o contenido. */
export class ValidationError extends AppError {
  readonly statusCode = 400;
}

/** Una matriz no tiene filas de longitud homogénea. */
export class InconsistentRowError extends ValidationError {
  constructor(matrixName: string, rowIndex: number, expected: number, actual: number) {
    super(
      `la fila ${rowIndex} de la matriz ${matrixName} tiene ${actual} columnas, ` +
        `se esperaban ${expected}: la matriz no es rectangular`,
    );
  }
}

/** Una matriz contiene un valor que no es un número finito. */
export class NonFiniteValueError extends ValidationError {
  constructor(matrixName: string, rowIndex: number, colIndex: number) {
    super(
      `la matriz ${matrixName} contiene un valor no finito en la posición ` +
        `[${rowIndex}][${colIndex}]`,
    );
  }
}

/** No se recibió ninguna matriz con la cual calcular. */
export class EmptyInputError extends ValidationError {
  constructor() {
    super('se requiere al menos una matriz con al menos un valor');
  }
}
```

### 6.3 Validación de dominio — `src/domain/validation.ts`

```typescript
import { InconsistentRowError, NonFiniteValueError } from './errors';
import type { NamedMatrix } from './matrix';

/**
 * Verifica los invariantes que Zod no cubre: que cada matriz sea
 * rectangular y que todos sus valores sean números finitos.
 *
 * Zod valida la forma del JSON; esto valida la coherencia matemática.
 * Son responsabilidades distintas y conviene mantenerlas separadas.
 */
export function assertValidMatrices(matrices: readonly NamedMatrix[]): void {
  for (const { name, data } of matrices) {
    const firstRow = data[0];
    if (firstRow === undefined) continue;

    const width = firstRow.length;

    for (let i = 0; i < data.length; i++) {
      const row = data[i];
      if (row === undefined) continue;

      if (row.length !== width) {
        throw new InconsistentRowError(name, i, width, row.length);
      }

      for (let j = 0; j < row.length; j++) {
        const value = row[j];
        if (value === undefined || !Number.isFinite(value)) {
          throw new NonFiniteValueError(name, i, j);
        }
      }
    }
  }
}
```

### 6.4 Servicio — `src/service/statistics.ts`

```typescript
import { EmptyInputError } from '../domain/errors';
import { flattenAll, isDiagonal, type NamedMatrix } from '../domain/matrix';

/** Las cinco métricas que exige el enunciado. */
export interface Statistics {
  readonly max: number;
  readonly min: number;
  readonly sum: number;
  readonly average: number;
  readonly isAnyDiagonal: boolean;
}

/**
 * Calcula las cinco estadísticas sobre el conjunto unificado de valores
 * de todas las matrices recibidas.
 *
 * Función pura: no accede a la petición HTTP ni produce efectos.
 * Máximo, mínimo y suma se obtienen en un único recorrido en lugar de tres.
 *
 * El promedio no se redondea: el redondeo es responsabilidad de la
 * capa de presentación, no del cálculo.
 */
export function calculateStatistics(matrices: readonly NamedMatrix[]): Statistics {
  const values = flattenAll(matrices);

  if (values.length === 0) {
    throw new EmptyInputError();
  }

  let max = -Infinity;
  let min = Infinity;
  let sum = 0;

  for (const value of values) {
    if (value > max) max = value;
    if (value < min) min = value;
    sum += value;
  }

  return {
    max,
    min,
    sum,
    average: sum / values.length,
    isAnyDiagonal: matrices.some(({ data }) => isDiagonal(data)),
  };
}
```

### 6.5 Esquema de entrada — `src/routes/statistics/schema.ts`

```typescript
import { z } from 'zod';

/**
 * Esquema del cuerpo de la petición.
 *
 * Valida únicamente la forma del JSON. La coherencia matemática
 * (filas homogéneas, valores finitos) la verifica el dominio.
 */
export const statisticsRequestSchema = z.object({
  matrices: z
    .array(
      z.object({
        name: z.string().min(1, 'el nombre de la matriz no puede estar vacío'),
        data: z
          .array(z.array(z.number()).min(1, 'una fila no puede estar vacía'))
          .min(1, 'una matriz no puede estar vacía'),
      }),
    )
    .min(1, 'se requiere al menos una matriz'),
});

/** El tipo se infiere del esquema: no puede divergir de la validación. */
export type StatisticsRequest = z.infer<typeof statisticsRequestSchema>;
```

### 6.6 Middleware de validación — `src/middleware/validateBody.ts`

```typescript
import type { NextFunction, Request, Response } from 'express';
import { ZodError, type ZodSchema } from 'zod';

import { ValidationError } from '../domain/errors';

/**
 * Valida el cuerpo de la petición contra un esquema Zod.
 *
 * Ante un fallo, traduce el error de Zod a un error del dominio y lo
 * delega a `next`, de modo que el manejador central sea el único
 * responsable de construir la respuesta.
 */
export function validateBody(schema: ZodSchema) {
  return (req: Request, _res: Response, next: NextFunction): void => {
    try {
      req.body = schema.parse(req.body);
      next();
    } catch (error) {
      if (error instanceof ZodError) {
        // `issues` es la propiedad estable: existe tanto en Zod 3 como en 4.
        // El alias `errors` de Zod 3 desapareció en la versión 4.
        const detail = error.issues
          .map((issue) => `${issue.path.join('.')}: ${issue.message}`)
          .join('; ');

        next(new ValidationError(`entrada inválida · ${detail}`));
        return;
      }

      next(error);
    }
  };
}
```

### 6.7 Manejador de errores — `src/middleware/errorHandler.ts`

```typescript
import type { NextFunction, Request, Response } from 'express';

import { AppError } from '../domain/errors';
import { config } from '../config';

/**
 * Único punto donde un error se convierte en respuesta HTTP.
 *
 * Los errores del dominio portan su propio código de estado. Cualquier
 * otro se trata como fallo interno y su detalle no se expone al cliente.
 *
 * El formato coincide con el de la API en Go, según el requerimiento
 * de coherencia estructural.
 */
export function errorHandler(
  error: Error,
  _req: Request,
  res: Response,
  _next: NextFunction,
): void {
  if (error instanceof AppError) {
    res.status(error.statusCode).json({
      success: false,
      error: error.message,
    });
    return;
  }

  if (config.isDevelopment) {
    console.error(error);
  }

  res.status(500).json({
    success: false,
    error: 'error interno al procesar la solicitud',
  });
}
```

### 6.8 Controlador — `src/routes/statistics/controller.ts`

```typescript
import type { NextFunction, Request, Response } from 'express';

import { assertValidMatrices } from '../../domain/validation';
import { calculateStatistics } from '../../service/statistics';
import type { StatisticsRequest } from './schema';

/**
 * Adaptador entre HTTP y el servicio de cálculo.
 *
 * No contiene lógica de negocio: extrae los datos, invoca al servicio y
 * responde. Los errores se delegan a `next` para que los maneje el
 * middleware central.
 *
 * La respuesta es el objeto de estadísticas sin envoltura: la API en Go
 * lo reenvía tal cual al cliente bajo la clave `statistics`.
 */
export function calculate(req: Request, res: Response, next: NextFunction): void {
  try {
    const { matrices } = req.body as StatisticsRequest;

    assertValidMatrices(matrices);

    res.status(200).json(calculateStatistics(matrices));
  } catch (error) {
    next(error);
  }
}
```

### 6.9 Rutas — `src/routes/statistics/index.ts` e `src/routes/index.ts`

```typescript
// src/routes/statistics/index.ts
import { Router } from 'express';

import { validateBody } from '../../middleware/validateBody';
import { calculate } from './controller';
import { statisticsRequestSchema } from './schema';

const router = Router();

router.post('/', validateBody(statisticsRequestSchema), calculate);

export default router;
```

```typescript
// src/routes/index.ts
import { Router } from 'express';

import statisticsRoutes from './statistics';

const router = Router();

router.use('/statistics', statisticsRoutes);

export default router;
```

### 6.10 Configuración — `src/config/index.ts`

```typescript
/**
 * Configuración leída del entorno, con valores por defecto aptos para
 * desarrollo local. Ningún valor sensible se versiona.
 */
export const config = {
  port: Number(process.env.PORT ?? 4000),
  nodeEnv: process.env.NODE_ENV ?? 'development',
  get isDevelopment(): boolean {
    return this.nodeEnv !== 'production';
  },
} as const;
```

### 6.11 Aplicación — `src/app.ts`

```typescript
import cors from 'cors';
import express, { type Express } from 'express';
import helmet from 'helmet';

import { errorHandler } from './middleware/errorHandler';
import { requestLogger } from './middleware/requestLogger';
import routes from './routes';

/**
 * Construye la aplicación Express sin ponerla a escuchar.
 *
 * Exportarla sin `listen` es lo que permite a Supertest ejercitar los
 * endpoints en las pruebas sin abrir un puerto real.
 */
export function createApp(): Express {
  const app = express();

  app.use(helmet());
  app.use(cors());
  app.use(express.json({ limit: '5mb' }));
  app.use(requestLogger);

  app.get('/health', (_req, res) => {
    res.json({ status: 'ok' });
  });

  app.use('/api/v1', routes);

  // El manejador de errores va al final: Express lo identifica por su aridad.
  app.use(errorHandler);

  return app;
}
```

### 6.12 Registro de peticiones — `src/middleware/requestLogger.ts`

```typescript
import type { NextFunction, Request, Response } from 'express';

/**
 * Registra método, ruta, código de estado y duración de cada petición.
 * Se suscribe a `finish` para medir el ciclo completo.
 */
export function requestLogger(req: Request, res: Response, next: NextFunction): void {
  const start = Date.now();

  res.on('finish', () => {
    const elapsed = Date.now() - start;
    console.log(`${req.method} ${req.originalUrl} ${res.statusCode} ${elapsed}ms`);
  });

  next();
}
```

### 6.13 Arranque — `src/server.ts`

```typescript
import { createApp } from './app';
import { config } from './config';

const app = createApp();

const server = app.listen(config.port, () => {
  console.log(`servicio de estadísticas escuchando en :${config.port}`);
});

/** Cierre ordenado ante señales del sistema, necesario en contenedores. */
const shutdown = (signal: string): void => {
  console.log(`${signal} recibido, cerrando servidor`);
  server.close(() => process.exit(0));
};

process.on('SIGTERM', () => shutdown('SIGTERM'));
process.on('SIGINT', () => shutdown('SIGINT'));
```

---

## 7. Pruebas

### 7.1 Dominio — `tests/unit/matrix.test.ts`

```typescript
import { cols, flattenAll, isDiagonal, rows } from '../../src/domain/matrix';

describe('isDiagonal', () => {
  it.each([
    ['diagonal simple', [[5, 0], [0, 3]], true],
    ['cero en la diagonal sigue siendo diagonal', [[5, 0], [0, 0]], true],
    ['valor fuera de la diagonal la invalida', [[5, 7], [0, 3]], false],
    ['no cuadrada con valor fuera de la diagonal', [[5, 4], [0, 3], [0, 0]], false],
    ['un solo elemento', [[7]], true],
    ['identidad 3x3', [[1, 0, 0], [0, 1, 0], [0, 0, 1]], true],
  ])('%s', (_caso, matriz, esperado) => {
    expect(isDiagonal(matriz as number[][])).toBe(esperado);
  });

  it('reconoce como no diagonal la matriz Q del caso de referencia', () => {
    const q = [
      [0.6, -0.8, 0],
      [0.8, 0.6, 0],
      [0, 0, 1],
    ];
    expect(isDiagonal(q)).toBe(false);
  });
});

describe('flattenAll', () => {
  it('une los valores de varias matrices en un solo arreglo', () => {
    const result = flattenAll([
      { name: 'A', data: [[1, 2], [3, 4]] },
      { name: 'B', data: [[5]] },
    ]);

    expect(result).toEqual([1, 2, 3, 4, 5]);
  });

  it('preserva la cantidad total de elementos', () => {
    const q = [[0.6, -0.8, 0], [0.8, 0.6, 0], [0, 0, 1]];
    const r = [[5, 4], [0, 3], [0, 0]];

    expect(flattenAll([
      { name: 'Q', data: q },
      { name: 'R', data: r },
    ])).toHaveLength(15);
  });
});

describe('dimensiones', () => {
  it('reporta filas y columnas', () => {
    const m = [[1, 2, 3], [4, 5, 6]];
    expect(rows(m)).toBe(2);
    expect(cols(m)).toBe(3);
  });

  it('reporta cero columnas para una matriz vacía', () => {
    expect(cols([])).toBe(0);
  });
});
```

### 7.2 Servicio — `tests/unit/statistics.test.ts`

```typescript
import { EmptyInputError } from '../../src/domain/errors';
import { calculateStatistics } from '../../src/service/statistics';

describe('calculateStatistics', () => {
  it('calcula las cinco métricas del conjunto de referencia', () => {
    const q = [
      [0.6, -0.8, 0],
      [0.8, 0.6, 0],
      [0, 0, 1],
    ];
    const r = [
      [5, 4],
      [0, 3],
      [0, 0],
    ];

    const stats = calculateStatistics([
      { name: 'Q', data: q },
      { name: 'R', data: r },
    ]);

    expect(stats.max).toBeCloseTo(5, 9);
    expect(stats.min).toBeCloseTo(-0.8, 9);
    expect(stats.sum).toBeCloseTo(14.2, 9);
    expect(stats.average).toBeCloseTo(14.2 / 15, 9);
    expect(stats.isAnyDiagonal).toBe(false);
  });

  it('maneja valores negativos', () => {
    const stats = calculateStatistics([{ name: 'A', data: [[-5, -2], [-9, -1]] }]);

    expect(stats.max).toBe(-1);
    expect(stats.min).toBe(-9);
    expect(stats.sum).toBe(-17);
  });

  it('maneja una matriz de un solo elemento', () => {
    const stats = calculateStatistics([{ name: 'A', data: [[7]] }]);

    expect(stats.max).toBe(7);
    expect(stats.min).toBe(7);
    expect(stats.sum).toBe(7);
    expect(stats.average).toBe(7);
    expect(stats.isAnyDiagonal).toBe(true);
  });

  it('incluye los ceros en el conteo del promedio', () => {
    const stats = calculateStatistics([{ name: 'A', data: [[0, 0], [0, 4]] }]);

    expect(stats.average).toBe(1); // 4 / 4, no 4 / 1
  });

  it('detecta la diagonal cuando al menos una matriz lo es', () => {
    const stats = calculateStatistics([
      { name: 'A', data: [[1, 2], [3, 4]] },
      { name: 'B', data: [[5, 0], [0, 6]] },
    ]);

    expect(stats.isAnyDiagonal).toBe(true);
  });

  it('rechaza un conjunto sin valores', () => {
    expect(() => calculateStatistics([])).toThrow(EmptyInputError);
  });
});
```

### 7.3 Integración — `tests/integration/statistics.route.test.ts`

```typescript
import request from 'supertest';

import { createApp } from '../../src/app';

const app = createApp();

describe('POST /api/v1/statistics', () => {
  it('devuelve las cinco métricas del conjunto de referencia', async () => {
    const response = await request(app)
      .post('/api/v1/statistics')
      .send({
        matrices: [
          { name: 'Q', data: [[0.6, -0.8, 0], [0.8, 0.6, 0], [0, 0, 1]] },
          { name: 'R', data: [[5, 4], [0, 3], [0, 0]] },
        ],
      })
      .expect(200);

    expect(response.body.max).toBeCloseTo(5, 9);
    expect(response.body.min).toBeCloseTo(-0.8, 9);
    expect(response.body.sum).toBeCloseTo(14.2, 9);
    expect(response.body.isAnyDiagonal).toBe(false);
  });

  it('responde sin envoltura, tal como lo espera la API en Go', async () => {
    const response = await request(app)
      .post('/api/v1/statistics')
      .send({ matrices: [{ name: 'A', data: [[1]] }] })
      .expect(200);

    expect(Object.keys(response.body).sort()).toEqual(
      ['average', 'isAnyDiagonal', 'max', 'min', 'sum'],
    );
  });

  it('rechaza un cuerpo sin matrices', async () => {
    const response = await request(app)
      .post('/api/v1/statistics')
      .send({})
      .expect(400);

    expect(response.body.success).toBe(false);
    expect(typeof response.body.error).toBe('string');
  });

  it('rechaza filas de longitudes distintas', async () => {
    const response = await request(app)
      .post('/api/v1/statistics')
      .send({ matrices: [{ name: 'Q', data: [[1, 2], [3]] }] })
      .expect(400);

    expect(response.body.error).toContain('rectangular');
  });

  it('rechaza valores no numéricos', async () => {
    await request(app)
      .post('/api/v1/statistics')
      .send({ matrices: [{ name: 'Q', data: [['a', 2]] }] })
      .expect(400);
  });
});

describe('GET /health', () => {
  it('reporta el servicio disponible', async () => {
    const response = await request(app).get('/health').expect(200);
    expect(response.body).toEqual({ status: 'ok' });
  });
});
```

---

## 8. Contenedor

```dockerfile
# ---- etapa de compilación ----
FROM node:20-alpine AS builder

WORKDIR /build

# Las dependencias se copian primero para aprovechar la caché de capas.
COPY package*.json ./
RUN npm ci

COPY tsconfig.json ./
COPY src ./src

RUN npm run build

# ---- etapa de dependencias de producción ----
FROM node:20-alpine AS deps

WORKDIR /deps

COPY package*.json ./
RUN npm ci --omit=dev

# ---- etapa final ----
FROM node:20-alpine

RUN addgroup -S app && adduser -S -G app app

WORKDIR /app

COPY --from=deps  /deps/node_modules ./node_modules
COPY --from=builder /build/dist ./dist
COPY package.json ./

USER app

EXPOSE 4000

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s \
    CMD wget -qO- http://localhost:4000/health || exit 1

CMD ["node", "dist/server.js"]
```

La imagen final no contiene TypeScript, ni las dependencias de desarrollo, ni el código fuente: solo `dist/` y las dependencias de producción.

---

## 9. Puntos de defensa para la entrevista

1. **La respuesta no lleva envoltura por una razón concreta.** `fiber-api` la reenvía tal cual bajo `statistics`; envolverla obligaría a Go a desenvolverla, añadiendo acoplamiento sin beneficio. Los errores sí van envueltos, en el mismo formato que la API en Go.

2. **La aplicación se exporta separada del servidor.** `app.ts` no llama a `listen`; `server.ts` sí. Eso permite probar con Supertest sin abrir puertos, lo que hace la suite rápida y sin conflictos.

3. **La validación está dividida en dos responsabilidades.** Zod valida la forma del JSON en el borde; el dominio valida la coherencia matemática (filas homogéneas, valores finitos). Son problemas distintos y conviene que fallen por separado, con mensajes distintos.

4. **El tipo se infiere del esquema Zod.** Al usar `z.infer` no existe una segunda declaración del tipo que pueda divergir de la validación.

5. **Un solo recorrido para tres métricas.** Máximo, mínimo y suma se obtienen en una pasada en lugar de tres, según el criterio de eficiencia del enunciado.

6. **`noUncheckedIndexedAccess` activado.** Obliga a comprobar cada acceso por índice. Es más verboso, pero elimina por construcción la clase de error más común al recorrer matrices.

7. **El servicio no conoce el origen de las matrices.** No sabe que provienen de una factorización QR. Esa ignorancia lo mantiene reutilizable y testeable de forma aislada.

8. **La arquitectura espeja la de la API en Go.** `domain` / `service` / capa de transporte, con las dependencias apuntando hacia adentro, cumpliendo el requerimiento de coherencia estructural entre ambos servicios.
