const request = require('supertest');
const app = require('../index');

describe('API basic endpoints', () => {
  test('GET /api returns health message', async () => {
    const res = await request(app).get('/api');

    expect(res.statusCode).toBe(200);
    expect(res.text).toBe('API is running');
  });

  test('GET /metrics returns Prometheus metrics', async () => {
    const res = await request(app).get('/metrics');

    expect(res.statusCode).toBe(200);
    expect(res.headers['content-type']).toContain('text/plain');
    expect(res.text).toContain('http_requests_total');
  });

  test('POST /api/jobs without description returns 400', async () => {
    const res = await request(app)
      .post('/api/jobs')
      .send({}); // no description

    expect(res.statusCode).toBe(400);
    expect(res.body).toEqual({ error: 'Description is required' });
  });
});