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
