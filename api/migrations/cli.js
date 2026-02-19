const { Pool } = require('pg');
const { runMigrations } = require('./migrate');

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
    } catch (err) {
      console.log('Host:', process.env.POSTGRES_HOST);
      console.log('PORT:', process.env.POSTGRES_PORT);
      console.log('Waiting for Postgres...', err.message);
      await new Promise(r => setTimeout(r, 2000));
      attempts++;
    }
  }
  throw new Error('Postgres did not become ready in time');
}

(async () => {
  try {
    await waitForDB();
    await runMigrations(pool, {
      enableCron: process.env.cron_enabled === 'true',
    });
    console.log('Migrations complete');
    process.exit(0);
  } catch (err) {
    console.error('Migration failed', err);
    process.exit(1);
  }
})();
