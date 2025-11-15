class HealthCheckUseCase {
  constructor(repository) {
    this.repository = repository;
  }

  async execute() {
    await this.repository.health();
  }
}

module.exports = HealthCheckUseCase;
