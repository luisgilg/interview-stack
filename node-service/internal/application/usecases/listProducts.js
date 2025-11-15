const { toResponseList } = require('../dto/productDto');

class ListProductsUseCase {
  constructor(repository, logger, cacheService) {
    this.repository = repository;
    this.logger = logger;
    this.cache = cacheService;
  }

  async execute() {
    const loader = async () => {
      const products = await this.repository.list();
      return toResponseList(products);
    };

    const { value, status } = await this.cache.fetch('products:list', loader);
    this.annotateCacheStatus(value, status);
    this.logger.info('listed products', { count: value.length, cacheStatus: status });
    return value;
  }

  annotateCacheStatus(products, status) {
    if (!status || !Array.isArray(products)) {
      return;
    }
    for (const product of products) {
      product.cache_status = status;
    }
  }
}

module.exports = ListProductsUseCase;
