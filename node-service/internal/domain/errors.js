class DomainError extends Error {
  /**
   * @param {'validation'|'not_found'|'internal'} code
   * @param {string} message
   * @param {Error} [cause]
   */
  constructor(code, message, cause) {
    super(message);
    this.name = 'DomainError';
    this.code = code;
    this.cause = cause;
  }

  static validation(message) {
    return new DomainError('validation', message);
  }

  static notFound(message) {
    return new DomainError('not_found', message);
  }

  static internal(message, cause) {
    return new DomainError('internal', message, cause);
  }
}

module.exports = DomainError;
