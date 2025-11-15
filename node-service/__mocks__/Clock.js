const { jest } = require('@jest/globals');

class ClockMock {
  constructor(now = new Date()) {
    this.current = new Date(now);
    this.now = jest.fn(() => new Date(this.current));
  }

  set(now) {
    this.current = new Date(now);
  }
}

module.exports = ClockMock;
