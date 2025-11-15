const DomainError = require('../../domain/errors');
const { toResponse } = require('../dto/productDto');

class GetProductUseCase {
  constructor(repository, logger, cacheService) {
    this.repository = repository;
    this.logger = logger;
    this.cache = cacheService;
  }

  async execute(id) {
    const loader = async () => {
      const product = await this.repository.getById(id);
      if (!product) {
        throw DomainError.notFound('product not found');
      }
      return toResponse(product);
    };

    const { value, status } = await this.cache.fetch(`products:${id}`, loader);
    if (value) {
      value.cache_status = status;
    }
    this.logger.info('product retrieved', { id, cacheStatus: status });
    return value;
  }
}

module.exports = GetProductUseCase;
