const request = require('supertest');
const { createApp } = require('../app');

jest.mock('pg', () => {
  const mClient = {
    query: jest.fn(),
    release: jest.fn(),
  };
  const mPool = {
    connect: jest.fn(() => Promise.resolve(mClient)),
    query: jest.fn(),
  };
  return { Pool: jest.fn(() => mPool) };
});

const { Pool } = require('pg');

describe('Job API DB integration (mocked)', () => {
  let pool;
  let client;

  beforeEach(() => {
    pool = new Pool();
    client = {
      query: jest.fn(),
      release: jest.fn(),
    };
    pool.connect.mockResolvedValue(client);
    app = createApp({ pool });
  });

  test('POST /api/jobs with description inserts job and returns 201', async () => {
    const fakeJob = { id: 1, description: 'Test job' };

    // Simulate the DB returning the inserted job
    client.query.mockImplementation((sql, params) => {
      if (sql.startsWith('BEGIN') || sql.startsWith('COMMIT')) {
        return Promise.resolve();
      }
      if (sql.startsWith('INSERT INTO jobs')) {
        return Promise.resolve({ rows: [fakeJob] });
      }
    });

    const res = await request(app)
      .post('/api/jobs')
      .send({ description: 'Test job' });

    expect(res.statusCode).toBe(201);
    expect(res.body).toEqual(fakeJob);

    // Check that BEGIN, INSERT, COMMIT were called
    expect(client.query).toHaveBeenCalledWith('BEGIN');
    expect(client.query).toHaveBeenCalledWith(
      'INSERT INTO jobs(description) VALUES($1) RETURNING id, description',
      ['Test job']
    );
    expect(client.query).toHaveBeenCalledWith('COMMIT');
    expect(client.release).toHaveBeenCalled();
  });

  test('POST /api/jobs DB failure triggers rollback and 500', async () => {
    jest.spyOn(console, 'error').mockImplementation(() => {});
    const error = new Error('DB error');

    client.query.mockImplementation((sql) => {
      if (sql.startsWith('BEGIN')) return Promise.resolve();
      if (sql.startsWith('INSERT INTO jobs')) throw error;
      if (sql.startsWith('ROLLBACK')) return Promise.resolve();
    });

    const res = await request(app)
      .post('/api/jobs')
      .send({ description: 'Fail job' });

    expect(res.statusCode).toBe(500);
    expect(res.body).toEqual({ error: 'Database error' });
    expect(client.query).toHaveBeenCalledWith('ROLLBACK');
    expect(client.release).toHaveBeenCalled();
    console.error.mockRestore();
  });
});
