import type { NextFunction, Request, Response } from 'express';

import { assertValidMatrices } from '../../domain/validation';
import { calculateStatistics } from '../../service/statistics';
import type { StatisticsRequest } from './schema';

/*
 * Adaptador entre HTTP y el servicio de cálculo.
 *
 * La respuesta es el objeto de estadísticas sin envolverlo: la API en Go
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
