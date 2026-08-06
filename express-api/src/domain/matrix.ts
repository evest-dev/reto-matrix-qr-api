/** Matriz rectangular por convención: el tipo no lo garantiza, assertValidMatrices sí. */
export type Matrix = readonly (readonly number[])[];

/** Matriz identificada por nombre, tal como llega desde la API de factorización. */
export interface NamedMatrix {
  readonly name: string;
  readonly data: Matrix;
}

export function rows(matrix: Matrix): number {
  return matrix.length;
}

/** 0 para una matriz vacía. */
export function cols(matrix: Matrix): number {
  return matrix[0]?.length ?? 0;
}

/**
 * Diagonal cuando todo elemento fuera de la diagonal principal es cero.
 * Sobre la diagonal vale cualquier valor, incluido cero: `[[5,0],[0,0]]`
 * sigue siendo diagonal. Aplica igual a matrices no cuadradas.
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

/** El enunciado pide "todos los valores de las matrices", en plural: se tratan como un solo conjunto. */
export function flattenAll(matrices: readonly NamedMatrix[]): number[] {
  const values: number[] = [];

  for (const { data } of matrices) {
    for (const row of data) {
      values.push(...row);
    }
  }

  return values;
}
