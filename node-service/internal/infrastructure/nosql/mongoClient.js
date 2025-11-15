const mongoose = require('mongoose');

async function connect(mongoConfig) {
  mongoose.set('strictQuery', true);
  await mongoose.connect(mongoConfig.uri, {
    dbName: mongoConfig.database,
    serverSelectionTimeoutMS: mongoConfig.connectTimeout,
    connectTimeoutMS: mongoConfig.connectTimeout,
    socketTimeoutMS: mongoConfig.operationTimeout
  });
}

async function disconnect() {
  if (mongoose.connection.readyState !== 0) {
    await mongoose.disconnect();
  }
}

async function health() {
  if (mongoose.connection.readyState !== 1) {
    throw new Error('mongo not connected');
  }
}

module.exports = {
  connect,
  disconnect,
  health
};
