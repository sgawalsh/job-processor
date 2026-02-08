const express = require('express');
const jobsRoutes = require('./routes/jobs');
const {
  metricsMiddleware,
  metricsEndpoint,
} = require('./metrics');

function createApp({ pool }) {
  const app = express();
  app.use(express.json());
  
  app.use(metricsMiddleware);
  app.use('/api/jobs', jobsRoutes({ pool }));
  app.get('/metrics', metricsEndpoint);

  app.get('/api', (_req, res) => {
    res.send('API is running');
  });

  return app;
}

module.exports = { createApp };