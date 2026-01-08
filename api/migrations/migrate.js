const { Pool } = require('pg');

// Use same env vars
const pool = new Pool({
  host: process.env.DB_HOST,
  port: process.env.DB_PORT,
  database: process.env.DB_NAME,
  user: process.env.DB_USER,
  password: process.env.DB_PASSWORD,
});

async function waitForDB() {
  let attempts = 0;
  while (attempts < 20) {
    try {
      await pool.query('SELECT 1');
      return;
    } catch {
      console.log('Waiting for Postgres...');
      await new Promise(r => setTimeout(r, 2000));
      attempts++;
    }
  }
  throw new Error('Postgres did not become ready in time');
}

async function runMigrations() {
  try {
    waitForDB();
    // Create enum type for job status
    await pool.query(`
      DO $$
      BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'job_status') THEN
          CREATE TYPE job_status AS ENUM (
            'PENDING',
            'QUEUED',
            'RUNNING',
            'SUCCEEDED',
            'FAILED'
          );
        END IF;
      END$$;
    `);
    // Create jobs table
    await pool.query(`
      CREATE TABLE IF NOT EXISTS jobs (
        id SERIAL PRIMARY KEY,
        description TEXT NOT NULL,
        status job_status NOT NULL DEFAULT 'PENDING',
        created_at TIMESTAMP DEFAULT NOW(),
        updated_at TIMESTAMP DEFAULT NOW(),
        enqueued_at TIMESTAMP,
        started_at TIMESTAMP,
        attempts INT DEFAULT 0,
        last_error TEXT
      )
    `);
    // Create pg_cron extension
    await pool.query(`CREATE EXTENSION IF NOT EXISTS pg_cron;`);

    await pool.query(`
      DO $$
      BEGIN
        -- Remove existing job if present
        PERFORM cron.unschedule(jobid)
        FROM cron.job
        WHERE jobname = 'cleanup_jobs';

        -- (Re)create job
        PERFORM cron.schedule(
          'cleanup_jobs',
          '0 * * * *',
          $exec$
            DELETE FROM jobs
            WHERE status = 'SUCCEEDED'
              AND created_at < NOW() - INTERVAL '1 day';
          $exec$
        );
      END
      $$;
    `);

    console.log('Migrations complete');
    process.exit(0);
  } catch (err) {
    console.error('Migration failed', err);
    process.exit(1);
  }
}

runMigrations();