const { Pool } = require('pg');

// Use same env vars
const pool = new Pool({
  host: process.env.POSTGRES_HOST,
  port: process.env.POSTGRES_PORT,
  database: process.env.POSTGRES_DB,
  user: process.env.POSTGRES_USER,
  password: process.env.POSTGRES_PASSWORD,
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
    await waitForDB();
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
    if (process.env.cron_enabled === 'true') {
      // Ensure pg_cron extension exists
      await pool.query(`CREATE EXTENSION IF NOT EXISTS pg_cron;`);

      const jobName = 'cleanup_jobs';
      const schedule = process.env.cron_schedule || '0 * * * *';
      const retention = process.env.cron_job_retention || '1 day';
      const sqlCommand = `
        DELETE FROM jobs
        WHERE status = 'SUCCEEDED'
          AND created_at < NOW() - INTERVAL '${retention}';
      `;

      // Unschedule existing job (if any)
      await pool.query(`
        SELECT cron.unschedule(jobid)
        FROM cron.job
        WHERE jobname = $1;
      `, [jobName]);

      // Schedule new job
      await pool.query(`
        SELECT cron.schedule($1, $2, $3);
      `, [jobName, schedule, sqlCommand]);
    }

    console.log('Migrations complete');
    process.exit(0);
  } catch (err) {
    console.error('Migration failed', err);
    process.exit(1);
  }
}

runMigrations();