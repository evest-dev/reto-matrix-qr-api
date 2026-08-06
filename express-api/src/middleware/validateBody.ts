import type { NextFunction, Request, Response } from 'express';
import { ZodError, type ZodSchema } from 'zod';

import { ValidationError } from '../domain/errors';

/*
 * Valida y Traduce errores de Zod a errores de dominio y delega a 
 * `next`: quien es el manejador central que arma la respuesta. 
*/
export function validateBody(schema: ZodSchema) {
  return (req: Request, _res: Response, next: NextFunction): void => {
    try {
      req.body = schema.parse(req.body);
      next();
    } catch (error) {
      if (error instanceof ZodError) {

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
