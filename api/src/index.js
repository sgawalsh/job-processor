const express = require('express');
const { Pool } = require('pg');
const Redis = require('ioredis');
const redis = new Redis(process.env.REDIS_URL);

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
    
    const job = result.rows[0];
    await redis.lpush('jobs:queue', job.id);

    res.status(201).json(result.rows[0]);
  } catch (err) {
    await client.query('ROLLBACK');
    console.error(err);
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