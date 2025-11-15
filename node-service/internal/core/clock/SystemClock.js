const Clock = require('./Clock');

class SystemClock extends Clock {
  now() {
    return new Date();
  }
}

module.exports = SystemClock;
