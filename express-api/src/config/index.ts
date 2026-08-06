
/** Valores leídos del entorno, con default para desarrollo local. */

export const config = {
  port: Number(process.env.PORT ?? 4000),
  nodeEnv: process.env.NODE_ENV ?? 'development',
  get isDevelopment(): boolean {
    return this.nodeEnv !== 'production';
  },
} as const;
