const DomainError = require('../../domain/errors');

function mapError(err, res) {
  if (err instanceof DomainError) {
    if (err.code === 'validation') {
      return res.status(400).json({ error: err.message });
    }
    if (err.code === 'not_found') {
      return res.status(404).json({ error: err.message });
    }
    return res.status(500).json({ error: err.message });
  }

  if (err && err.statusCode) {
    return res.status(err.statusCode).json({ error: err.message });
  }

  return res.status(500).json({ error: 'internal server error' });
}

module.exports = mapError;
