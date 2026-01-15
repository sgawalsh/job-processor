const express = require('express');
const { Pool } = require('pg');
const clientProm = require('prom-client');

// Read database config from environment variables
const pool = new Pool({
  host: process.env.DB_HOST,
  port: process.env.DB_PORT,
  database: process.env.DB_NAME,
  user: process.env.DB_USER,
  password: process.env.DB_PASSWORD,
});

const app = express();
app.use(express.json()); // for parsing application/json

// -----------------------------
// Prometheus metrics setup
// -----------------------------

// Collect default Node.js metrics (CPU, memory, event loop, etc.)
clientProm.collectDefaultMetrics();

// Counter: total HTTP requests received
const httpRequestsTotal = new clientProm.Counter({
  name: 'http_requests_total',
  help: 'Total number of HTTP requests',
  labelNames: ['method', 'route', 'status'],
});

// Counter: total job insert failures
const jobFailures = new clientProm.Counter({
  name: 'jobs_failed_total',
  help: 'Total number of failed job inserts',
});

// Histogram: HTTP request duration in seconds
const httpRequestDuration = new clientProm.Histogram({
  name: 'http_request_duration_seconds',
  help: 'Duration of HTTP requests in seconds',
  labelNames: ['method', 'route', 'status'],
  buckets: [0.05, 0.1, 0.3, 0.5, 1, 2, 5], // adjust as needed
});

// -----------------------------
// Middleware to track metrics
// -----------------------------
app.use((req, res, next) => {
  const end = httpRequestDuration.startTimer({ method: req.method });
  res.on('finish', () => {
    const route = req.route?.path || req.path || 'unknown';
    const status = res.statusCode;
    httpRequestsTotal.inc({ method: req.method, route, status });
    end({ method: req.method, route, status });
  });
  next();
});

// Expose /metrics endpoint
app.get('/metrics', async (req, res) => {
  res.set('Content-Type', clientProm.register.contentType);
  res.end(await clientProm.register.metrics());
});

// -----------------------------
// API routes
// -----------------------------

// Health check
app.get('/', (req, res) => {
  res.send('API is running');
});

// Endpoint to create a job
app.post('/jobs', async (req, res) => {
  const { description } = req.body;
  if (!description) {
    return res.status(400).json({ error: 'description is required' });
  }

  const client = await pool.connect();

  try {
    await client.query('BEGIN');
    const result = await client.query(
      'INSERT INTO jobs(description) VALUES($1) RETURNING id, description',
      [description]
    );

    await client.query('COMMIT');
    res.status(201).json(result.rows[0]);
  } catch (err) {
    await client.query('ROLLBACK');
    console.error(err);
    jobFailures.inc();
    res.status(500).json({ error: 'Database error' });
  } finally {
    client.release();
  }
});

// Start server
const port = process.env.PORT || 8080;
app.listen(port, () => {
  console.log(`API running on port ${port}`);
});