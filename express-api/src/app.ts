import cors from 'cors';
import express, { type Express } from 'express';
import helmet from 'helmet';

import { errorHandler } from './middleware/errorHandler';
import { requestLogger } from './middleware/requestLogger';
import routes from './routes';

/**
 * Construye la aplicación Express sin ponerla a escuchar.
 *
 * Exporta sin `listen`: así Supertest prueba los endpoints sin abrir un puerto real.
*/

export function createApp(): Express {
  const app = express();

  app.use(helmet());
  app.use(cors());
  app.use(express.json({ limit: '5mb' }));
  app.use(requestLogger);

  app.get('/health', (_req, res) => {
    res.json({ status: 'ok' });
  });

  app.use('/api/v1', routes);

  // El manejador de errores va al final: Express lo identifica por su aridad.
  app.use(errorHandler);

  return app;
}
