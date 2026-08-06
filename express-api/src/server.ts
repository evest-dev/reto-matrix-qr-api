import { createApp } from './app';
import { config } from './config';

const app = createApp();

const server = app.listen(config.port, () => {
  console.log(`servicio de estadísticas escuchando en :${config.port}`);
});

/** Cierre ordenado ante señales del sistema, necesario en contenedores. */
const shutdown = (signal: string): void => {
  console.log(`${signal} recibido, cerrando servidor`);
  server.close(() => process.exit(0));
};

process.on('SIGTERM', () => shutdown('SIGTERM'));
process.on('SIGINT', () => shutdown('SIGINT'));
