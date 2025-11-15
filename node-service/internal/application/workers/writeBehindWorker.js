const Product = require('../../domain/product');
const metrics = require('../../observability/metrics');

function delay(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

class WriteBehindWorker {
  constructor({ client, streamName, group, consumer, batchSize, blockMs, store, logger, source }) {
    this.client = client;
    this.streamName = streamName;
    this.group = group;
    this.consumer = consumer;
    this.batchSize = batchSize;
    this.blockMs = blockMs;
    this.store = store;
    this.logger = logger;
    this.source = source;
    this.running = false;
    this.loopPromise = null;
  }

  async start() {
    if (this.running || !this.client) {
      return;
    }
    await this.ensureGroup();
    this.running = true;
    this.loopPromise = this.loop();
  }

  async stop() {
    this.running = false;
    if (this.loopPromise) {
      await this.loopPromise.catch(() => {});
    }
  }

  async ensureGroup() {
    try {
      await this.client.xGroupCreate(this.streamName, this.group, '0', { MKSTREAM: true });
    } catch (err) {
      if (!err?.message?.includes('BUSYGROUP')) {
        this.logger?.warn?.('failed to create consumer group', { error: err.message });
      }
    }
  }

  async loop() {
    while (this.running) {
      try {
        const response = await this.client.xReadGroup(
          this.group,
          this.consumer,
          { key: this.streamName, id: '>' },
          { COUNT: this.batchSize, BLOCK: this.blockMs }
        );
        if (!response) {
          continue;
        }
        await this.processStreams(response);
      } catch (err) {
        if (!this.running) {
          return;
        }
        this.logger?.warn?.('write-behind read failed', { error: err.message });
        metrics.recordWriteBehindError();
        await delay(1000);
      }
    }
  }

  async processStreams(streams) {
    const batchStart = process.hrtime.bigint();
    const now = Date.now();
    let processed = 0;
    let maxLagSeconds = 0;
    for (const stream of streams) {
      for (const message of stream.messages) {
        processed += 1;
        const lagSeconds = this.calculateLagSeconds(now, message.id);
        if (lagSeconds > maxLagSeconds) {
          maxLagSeconds = lagSeconds;
        }
        await this.handleMessage(message);
      }
    }
    if (processed === 0) {
      return;
    }
    let queueLength = -1;
    try {
      queueLength = await this.client.xLen(this.streamName);
    } catch (err) {
      this.logger?.warn?.('failed to read queue length', { error: err.message });
      metrics.recordWriteBehindError();
    }
    const durationSeconds = Number(process.hrtime.bigint() - batchStart) / 1e9;
    metrics.recordWriteBehindBatch({
      size: processed,
      durationSeconds,
      lagSeconds: maxLagSeconds,
      queueLength
    });
  }

  async handleMessage(message) {
    try {
      const envelope = message.message || {};
      const rawEvent = envelope.event;
      if (!rawEvent) {
        await this.ack(message.id);
        return;
      }
      const event = JSON.parse(rawEvent);
      if (event.source && event.source !== this.source) {
        await this.ack(message.id);
        return;
      }
      await this.applyEvent(event);
      await this.ack(message.id);
    } catch (err) {
      this.logger?.warn?.('write-behind apply failed', { error: err.message });
      metrics.recordWriteBehindError();
    }
  }

  async ack(id) {
    try {
      await this.client.xAck(this.streamName, this.group, id);
    } catch (err) {
      this.logger?.warn?.('failed to ack message', { id, error: err.message });
      metrics.recordWriteBehindError();
    }
  }

  async applyEvent(event) {
    switch (event.type) {
      case 'create':
        await this.applyCreate(event);
        break;
      case 'update':
        await this.applyUpdate(event);
        break;
      case 'delete':
        await this.applyDelete(event);
        break;
      default:
        this.logger?.warn?.('unknown write event type', { type: event.type });
    }
  }

  hydrateProduct(payload) {
    if (!payload) {
      return null;
    }
    const hydrated = {
      ...payload,
      createdAt: payload.createdAt ? new Date(payload.createdAt) : undefined,
      updatedAt: payload.updatedAt ? new Date(payload.updatedAt) : undefined
    };
    return new Product(hydrated);
  }

  async applyCreate(event) {
    const product = this.hydrateProduct(event.payload);
    if (!product) {
      throw new Error('missing payload for create event');
    }
    await this.store.createProduct(product);
  }

  async applyUpdate(event) {
    const product = this.hydrateProduct(event.payload);
    if (!product) {
      throw new Error('missing payload for update event');
    }
    await this.store.updateProduct(event.id, product);
  }

  async applyDelete(event) {
    await this.store.deleteProduct(event.id);
  }

  calculateLagSeconds(nowMs, id) {
    if (!id) {
      return 0;
    }
    const [timestamp] = String(id).split('-');
    const producedAt = Number(timestamp);
    if (!Number.isFinite(producedAt)) {
      return 0;
    }
    const lagMs = Math.max(0, nowMs - producedAt);
    return lagMs / 1000;
  }
}

module.exports = WriteBehindWorker;
