const express = require('express');
const swaggerUi = require('swagger-ui-express');

const { swaggerSpec } = require('./swagger');
const metrics = require('../../observability/metrics');

function createApp(config, controller) {
  const app = express();
  app.use(express.json());
  app.use(metrics.httpMiddleware());
  app.use((req, res, next) => {
    req.setTimeout(config.server.readTimeout);
    res.setTimeout(config.server.writeTimeout);
    next();
  });
  app.use('/swagger', swaggerUi.serve, swaggerUi.setup(swaggerSpec));
  app.get('/swagger.json', (_req, res) => {
    res.json(swaggerSpec);
  });
  controller.register(app);
  return app;
}

module.exports = {
  createApp
};
