/** Todo error esperado desciende de esta clase; instanceof la identifica en errorHandler. */
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
