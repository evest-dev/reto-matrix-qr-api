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
