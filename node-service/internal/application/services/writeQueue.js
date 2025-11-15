class WriteQueueProducer {
  constructor(redisClient, streamName, logger = console) {
    this.client = redisClient;
    this.streamName = streamName;
    this.logger = logger;
  }

  async enqueue(event) {
    if (!this.client) {
      throw new Error('write queue client not configured');
    }
    const payload = JSON.stringify(event);
    try {
      await this.client.xAdd(this.streamName, '*', { event: payload });
    } catch (err) {
      this.logger?.error?.('failed to enqueue write event', err);
      throw err;
    }
  }
}

module.exports = WriteQueueProducer;
