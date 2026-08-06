import type { NextFunction, Request, Response } from 'express';

import { AppError } from '../domain/errors';
import { config } from '../config';

/**
 * Único punto donde un error se traduce a HTTP: los del dominio usan su
 * propio código, el resto cae a 500 sin exponer detalle.
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
