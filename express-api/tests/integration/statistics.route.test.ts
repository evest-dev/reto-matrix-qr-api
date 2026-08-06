import request from 'supertest';

import { createApp } from '../../src/app';

const app = createApp();

describe('POST /api/v1/statistics', () => {
  it('devuelve las cinco métricas del conjunto de referencia', async () => {
    const response = await request(app)
      .post('/api/v1/statistics')
      .send({
        matrices: [
          { name: 'Q', data: [[0.6, -0.8, 0], [0.8, 0.6, 0], [0, 0, 1]] },
          { name: 'R', data: [[5, 4], [0, 3], [0, 0]] },
        ],
      })
      .expect(200);

    expect(response.body.max).toBeCloseTo(5, 9);
    expect(response.body.min).toBeCloseTo(-0.8, 9);
    expect(response.body.sum).toBeCloseTo(14.2, 9);
    expect(response.body.isAnyDiagonal).toBe(false);
  });

  it('responde sin envoltura, tal como lo espera la API en Go', async () => {
    const response = await request(app)
      .post('/api/v1/statistics')
      .send({ matrices: [{ name: 'A', data: [[1]] }] })
      .expect(200);

    expect(Object.keys(response.body).sort()).toEqual(
      ['average', 'isAnyDiagonal', 'max', 'min', 'sum'],
    );
  });

  it('rechaza un cuerpo sin matrices', async () => {
    const response = await request(app)
      .post('/api/v1/statistics')
      .send({})
      .expect(400);

    expect(response.body.success).toBe(false);
    expect(typeof response.body.error).toBe('string');
  });

  it('rechaza filas de longitudes distintas', async () => {
    const response = await request(app)
      .post('/api/v1/statistics')
      .send({ matrices: [{ name: 'Q', data: [[1, 2], [3]] }] })
      .expect(400);

    expect(response.body.error).toContain('rectangular');
  });

  it('rechaza valores no numéricos', async () => {
    await request(app)
      .post('/api/v1/statistics')
      .send({ matrices: [{ name: 'Q', data: [['a', 2]] }] })
      .expect(400);
  });
});

describe('GET /health', () => {
  it('reporta el servicio disponible', async () => {
    const response = await request(app).get('/health').expect(200);
    expect(response.body).toEqual({ status: 'ok' });
  });
});
