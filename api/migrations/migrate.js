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
      public_id UUID NOT NULL DEFAULT gen_random_uuid() UNIQUE,
      description TEXT NOT NULL,
      status job_status NOT NULL DEFAULT 'PENDING',
      created_at TIMESTAMPTZ DEFAULT NOW(),
      updated_at TIMESTAMPTZ DEFAULT NOW(),
      enqueued_at TIMESTAMPTZ,
      started_at TIMESTAMPTZ,
      attempts INT NOT NULL DEFAULT 0,
      last_error TEXT
    )
  `);
  
  // Create index on status for efficient querying of pending jobs for KEDA
  await pool.query(`
    CREATE INDEX IF NOT EXISTS idx_jobs_status_pending
    ON jobs(status)
    WHERE status = 'PENDING';
  `);

  // Create trigger function for automatic update_at timestamp
  await pool.query(`
    CREATE OR REPLACE FUNCTION set_updated_at()
    RETURNS TRIGGER AS $$
    BEGIN
      NEW.updated_at = NOW();
      RETURN NEW;
    END;
    $$ LANGUAGE plpgsql;;
  `);

  // Create trigger function for automatic update_at timestamp
  await pool.query(`
    DO $$
    BEGIN
      IF NOT EXISTS (
        SELECT 1 FROM pg_trigger WHERE tgname = 'jobs_set_updated_at'
      ) THEN
        CREATE TRIGGER jobs_set_updated_at
        BEFORE UPDATE ON jobs
        FOR EACH ROW
        EXECUTE FUNCTION set_updated_at();
      END IF;
    END
    $$;
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

    console.log(`Scheduled cron job '${jobName}' with schedule '${schedule}' to clean up succeeded jobs older than ${retention}.`);
  }
}

module.exports = { runMigrations };