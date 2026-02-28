jest.setTimeout(30000); // allow container startup time

const { PostgreSqlContainer } = require('@testcontainers/postgresql');
const { Pool } = require('pg');
const request = require('supertest');
const { createApp } = require('../app');
const { runMigrations } = require('../../migrations/migrate');

let container;
let pool;
let app;

beforeAll(async () => {
  // Start the container
  container = await new PostgreSqlContainer('postgres:18-bookworm')
    .withDatabase('testdb')
    .withUsername('testuser')
    .withPassword('testpass')
    .start();

  // Connect the pool
  pool = new Pool({ connectionString: container.getConnectionUri() });
  app = createApp({ pool });

  await runMigrations(pool);
});

beforeEach(async () => {
  // Clean the table before each test
  await pool.query('TRUNCATE TABLE jobs RESTART IDENTITY');
});

afterAll(async () => {
  if (pool) await pool.end();
  if (container) await container.stop();
});

describe('POST /jobs integration', () => {
  test('creates a job successfully', async () => {
    const res = await request(app)
      .post('/jobs')
      .send({ description: 'Integration job' });

    expect(res.statusCode).toBe(201);
    expect(res.body).toHaveProperty('id');
    expect(res.body.description).toBe('Integration job');

    const dbRes = await pool.query('SELECT * FROM jobs WHERE id = $1', [res.body.id]);
    expect(dbRes.rowCount).toBe(1);
  });

  test('returns 400 if description missing', async () => {
    const res = await request(app).post('/jobs').send({});
    expect(res.statusCode).toBe(400);
    expect(res.body).toEqual({ error: 'Description is required' });
  });
});
