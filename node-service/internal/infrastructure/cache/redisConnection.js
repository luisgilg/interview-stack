const net = require('net');

function encodeCommand(args) {
  const parts = [Buffer.from(`*${args.length}\r\n`)];
  for (const arg of args) {
    const buffer = Buffer.isBuffer(arg) ? arg : Buffer.from(String(arg));
    parts.push(Buffer.from(`$${buffer.length}\r\n`));
    parts.push(buffer);
    parts.push(Buffer.from('\r\n'));
  }
  return Buffer.concat(parts);
}

function tryParse(buffer) {
  if (!buffer || buffer.length === 0) {
    return null;
  }
  const prefix = buffer[0];
  const lineEnd = buffer.indexOf('\r\n');
  if (lineEnd === -1) {
    return null;
  }
  const line = buffer.slice(1, lineEnd).toString();
  switch (prefix) {
    case 43: // '+'
      return { value: line, length: lineEnd + 2 };
    case 45: // '-'
      return { error: new Error(line), length: lineEnd + 2 };
    case 58: // ':'
      return { value: Number.parseInt(line, 10), length: lineEnd + 2 };
    case 36: { // '$'
      const length = Number.parseInt(line, 10);
      if (Number.isNaN(length)) {
        return { error: new Error('invalid bulk length'), length: lineEnd + 2 };
      }
      if (length === -1) {
        return { value: null, length: lineEnd + 2 };
      }
      const total = lineEnd + 2 + length + 2;
      if (buffer.length < total) {
        return null;
      }
      const value = buffer.slice(lineEnd + 2, lineEnd + 2 + length).toString();
      return { value, length: total };
    }
    default:
      return { error: new Error(`unsupported RESP type: ${String.fromCharCode(prefix)}`), length: lineEnd + 2 };
  }
}

class RedisConnection {
  constructor({ host, port }) {
    this.host = host;
    this.port = port;
    this.socket = null;
    this.buffer = Buffer.alloc(0);
    this.queue = [];
    this.connectPromise = null;
    this.closing = false;
  }

  connect() {
    if (this.connectPromise) {
      return this.connectPromise;
    }
    this.connectPromise = new Promise((resolve, reject) => {
      this.socket = net.createConnection({ host: this.host, port: this.port }, resolve);
      this.socket.on('data', (chunk) => {
        this.buffer = Buffer.concat([this.buffer, chunk]);
        this.flush();
      });
      this.socket.on('error', (err) => this.fail(err));
      this.socket.on('close', () => {
        if (!this.closing) {
          this.fail(new Error('redis connection closed'));
        }
      });
    });
    return this.connectPromise;
  }

  async send(args) {
    await this.connect();
    return new Promise((resolve, reject) => {
      this.queue.push({ resolve, reject });
      this.socket.write(encodeCommand(args));
    });
  }

  flush() {
    while (this.queue.length > 0) {
      const parsed = tryParse(this.buffer);
      if (!parsed) {
        break;
      }
      this.buffer = this.buffer.slice(parsed.length);
      const pending = this.queue.shift();
      if (parsed.error) {
        pending.reject(parsed.error);
      } else {
        pending.resolve(parsed.value);
      }
    }
  }

  fail(err) {
    while (this.queue.length > 0) {
      const pending = this.queue.shift();
      pending.reject(err);
    }
  }

  close() {
    this.closing = true;
    if (this.socket) {
      this.socket.removeAllListeners();
      this.socket.end();
      this.socket.destroy();
      this.socket = null;
    }
  }
}

module.exports = {
  RedisConnection
};
