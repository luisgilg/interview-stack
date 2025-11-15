const Clock = require('./Clock');

class FakeClock extends Clock {
  constructor(initialDate = new Date(0)) {
    super();
    this.current = new Date(initialDate);
  }

  now() {
    return new Date(this.current);
  }

  set(date) {
    this.current = new Date(date);
  }

  advance(milliseconds) {
    this.current = new Date(this.current.getTime() + milliseconds);
  }
}

module.exports = FakeClock;
