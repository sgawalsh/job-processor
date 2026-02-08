const clientProm = require('prom-client');

// Collect default Node.js metrics (CPU, memory, event loop, etc.)
clientProm.collectDefaultMetrics();

// Counter: total HTTP requests received
const httpRequestsTotal = new clientProm.Counter({
  name: 'http_requests_total',
  help: 'Total number of HTTP requests',
  labelNames: ['method', 'route', 'status'],
});

// Histogram: HTTP request duration in seconds
const httpRequestDuration = new clientProm.Histogram({
  name: 'http_request_duration_seconds',
  help: 'Duration of HTTP requests in seconds',
  labelNames: ['method', 'route', 'status'],
  buckets: [0.05, 0.1, 0.3, 0.5, 1, 2, 5], // adjust as needed
});

// Counter: total job insert failures
const jobFailures = new clientProm.Counter({
  name: 'api_jobs_failed_total',
  help: 'Total number of failed job inserts',
});

// -----------------------------
// Middleware to track metrics
// -----------------------------
function metricsMiddleware(req, res, next) {
  const end = httpRequestDuration.startTimer({ method: req.method });

  res.on('finish', () => {
    let route = 'unknown';

    if (req.route?.path) {
      route = req.baseUrl
        ? `${req.baseUrl}${req.route.path}`
        : req.route.path;
    }

    const status = res.statusCode;

    httpRequestsTotal.inc({ method: req.method, route, status });
    end({ method: req.method, route, status });
  });

  next();
}

// Expose /metrics endpoint
function metricsEndpoint(req, res) {
  res.set('Content-Type', clientProm.register.contentType);
  clientProm.register.metrics().then(metrics => res.end(metrics));
}

module.exports = {
  metricsMiddleware,
  metricsEndpoint,
  jobFailures,
};