const request = require('supertest');
const { createApp } = require('../app');

describe('API basic endpoints', () => {

  beforeAll(() => {
    const pool = {
      query: jest.fn(),
      connect: jest.fn(),
    };

    app = createApp({ pool });
  });

  test('GET /health returns health message', async () => {
    const res = await request(app).get('/health');

    expect(res.statusCode).toBe(200);
    expect(res.text).toBe('API is running');
  });

  test('GET /metrics returns Prometheus metrics', async () => {
    const res = await request(app).get('/metrics');

    expect(res.statusCode).toBe(200);
    expect(res.headers['content-type']).toContain('text/plain');
    expect(res.text).toContain('http_requests_total');
  });

  test('POST /jobs without description returns 400', async () => {
    const res = await request(app)
      .post('/jobs')
      .send({}); // no description

    expect(res.statusCode).toBe(400);
    expect(res.body).toEqual({ error: 'Description is required' });
  });
});