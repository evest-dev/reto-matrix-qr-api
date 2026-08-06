import type { NextFunction, Request, Response } from 'express';

/**
 * Registra método, ruta, código de estado y duración de cada petición.
 * Se engancha a `finish`, no al inicio, para medir el ciclo completo.
 */


export function requestLogger(req: Request, res: Response, next: NextFunction): void {
  const start = Date.now();

  res.on('finish', () => {
    const elapsed = Date.now() - start;
    console.log(`${req.method} ${req.originalUrl} ${res.statusCode} ${elapsed}ms`);
  });

  next();
}
