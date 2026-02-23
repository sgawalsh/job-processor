async function runMigrations(pool, { enableCron = false } = {}) {
  // Create enum type for job status
  await pool.query(`
    DO $$
    BEGIN
      IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'job_status') THEN
        CREATE TYPE job_status AS ENUM (
          'PENDING',
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
      started_at TIMESTAMP,
      attempts INT DEFAULT 0,
      last_error TEXT
    )
  `);

  await pool.query(`
    CREATE OR REPLACE FUNCTION notify_jobs_available()
    RETURNS trigger AS $$
    BEGIN
      IF NEW.status = 'PENDING'
        AND (TG_OP = 'INSERT' OR OLD.status IS DISTINCT FROM NEW.status)
      THEN
        PERFORM pg_notify('jobs_available', '');
      END IF;
      RETURN NEW;
    END;
    $$ LANGUAGE plpgsql;
  `);

  await pool.query(`
    DO $$
    BEGIN
      IF NOT EXISTS (
        SELECT 1
        FROM pg_trigger
        WHERE tgname = 'jobs_notify_trigger'
      ) THEN
        CREATE TRIGGER jobs_notify_trigger
        AFTER INSERT OR UPDATE OF status
        ON jobs
        FOR EACH ROW
        EXECUTE FUNCTION notify_jobs_available();
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