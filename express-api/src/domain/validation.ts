import { InconsistentRowError, NonFiniteValueError } from './errors';
import type { NamedMatrix } from './matrix';

/** Zod valida la forma del JSON; esto valida la coherencia matemática: filas rectangulares y valores finitos. */
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
