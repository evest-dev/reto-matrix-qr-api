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
 * Función pura: sin acceso a req/res. El promedio no se redondea — el
 * redondeo es responsabilidad de la capa de presentación, no del cálculo.
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
