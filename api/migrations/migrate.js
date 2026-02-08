async function runMigrations(pool, { enableCron = false } = {}) {
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
  if (enableCron) {
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
}

module.exports = { runMigrations };