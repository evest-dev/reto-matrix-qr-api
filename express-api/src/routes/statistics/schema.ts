import { z } from 'zod';

/** Valida solo la forma del JSON; la coherencia matemática la verifica el dominio. */
export const statisticsRequestSchema = z.object({
  matrices: z
    .array(
      z.object({
        name: z.string().min(1, 'el nombre de la matriz no puede estar vacío'),
        data: z
          .array(z.array(z.number()).min(1, 'una fila no puede estar vacía'))
          .min(1, 'una matriz no puede estar vacía'),
      }),
    )
    .min(1, 'se requiere al menos una matriz'),
});

/** El type se infiere del esquema: no puede divergir de la validación. */
export type StatisticsRequest = z.infer<typeof statisticsRequestSchema>;
