const { createApp } = require('./app');
const { createPool } = require('./db');

const pool = createPool();
const app = createApp({ pool });

const port = process.env.API_PORT || 8080;
app.listen(port, () => {
  console.log(`API running on port ${port}`);
});