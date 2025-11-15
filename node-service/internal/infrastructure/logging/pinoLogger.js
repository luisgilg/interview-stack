const pino = require('pino');

function createLogger() {
  const instance = pino({ level: process.env.LOG_LEVEL || 'info' });
  return {
    info: (msg, meta = {}) => instance.info(meta, msg),
    warn: (msg, meta = {}) => instance.warn(meta, msg),
    error: (msg, meta = {}) => instance.error(meta, msg)
  };
}

module.exports = {
  createLogger
};
