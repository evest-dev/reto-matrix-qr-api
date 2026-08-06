import type { NextFunction, Request, Response } from 'express';

/**
 * `config.nodeEnv` se fija una sola vez al cargar el módulo: para
 * ejercitar de verdad las dos ramas de `isDevelopment` hace falta
 * `jest.resetModules()` + `require` fresco, no un mock del getter.
 *
 * `ValidationError` se recarga junto con `errorHandler` por la misma
 * razón: instanceof falla tras resetModules, las clases quedan en
 * registros de módulos distintos.
 */
function loadModules(nodeEnv: string): {
  errorHandler: typeof import('../../src/middleware/errorHandler').errorHandler;
  ValidationError: typeof import('../../src/domain/errors').ValidationError;
} {
  jest.resetModules();
  process.env.NODE_ENV = nodeEnv;
  
  // eslint-disable-next-line @typescript-eslint/no-var-requires
  const { errorHandler } = require('../../src/middleware/errorHandler') as typeof import('../../src/middleware/errorHandler');
  
  // eslint-disable-next-line @typescript-eslint/no-var-requires
  const { ValidationError } = require('../../src/domain/errors') as typeof import('../../src/domain/errors');
  return { errorHandler, ValidationError };
}

function createMockResponse(): Response {
  const res = {} as Response;
  res.status = jest.fn().mockReturnValue(res);
  res.json = jest.fn().mockReturnValue(res);
  return res;
}

const req = {} as Request;
const next = jest.fn() as unknown as NextFunction;

describe('errorHandler', () => {
  const originalNodeEnv = process.env.NODE_ENV;

  afterEach(() => {
    jest.restoreAllMocks();
  });

  afterAll(() => {
    process.env.NODE_ENV = originalNodeEnv;
    jest.resetModules();
  });

  describe('con NODE_ENV=development', () => {
    const { errorHandler, ValidationError } = loadModules('development');

    it('un AppError responde con su statusCode y su mensaje', () => {
      const res = createMockResponse();
      const error = new ValidationError('mensaje de dominio');

      errorHandler(error, req, res, next);

      expect(res.status).toHaveBeenCalledWith(400);
      expect(res.json).toHaveBeenCalledWith({ success: false, error: 'mensaje de dominio' });
    });

    it('un error genérico responde 500 sin filtrar el mensaje, y se registra en consola', () => {
      jest.spyOn(console, 'error').mockImplementation(() => undefined);

      const res = createMockResponse();
      const error = new Error('detalle interno sensible');

      errorHandler(error, req, res, next);

      expect(res.status).toHaveBeenCalledWith(500);
      expect(res.json).toHaveBeenCalledWith({
        success: false,
        error: 'error interno al procesar la solicitud',
      });
      expect(console.error).toHaveBeenCalledWith(error);
    });
  });

  describe('con NODE_ENV=production', () => {
    const { errorHandler, ValidationError } = loadModules('production');

    it('un AppError responde con su statusCode y su mensaje', () => {
      const res = createMockResponse();
      const error = new ValidationError('mensaje de dominio');

      errorHandler(error, req, res, next);

      expect(res.status).toHaveBeenCalledWith(400);
      expect(res.json).toHaveBeenCalledWith({ success: false, error: 'mensaje de dominio' });
    });

    it('un error genérico responde 500 sin filtrar el mensaje, y no se registra en consola', () => {
      jest.spyOn(console, 'error').mockImplementation(() => undefined);

      const res = createMockResponse();
      const error = new Error('detalle interno sensible');

      errorHandler(error, req, res, next);

      expect(res.status).toHaveBeenCalledWith(500);
      expect(res.json).toHaveBeenCalledWith({
        success: false,
        error: 'error interno al procesar la solicitud',
      });
      expect(console.error).not.toHaveBeenCalled();
    });
  });
});
