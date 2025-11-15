module.exports = {
  clearMocks: true,
  testEnvironment: 'node',
  testMatch: ['**/__tests__/**/*.test.js'],
  roots: ['<rootDir>/__tests__'],
  moduleFileExtensions: ['js', 'json'],
  collectCoverageFrom: ['internal/**/*.js'],
  moduleNameMapper: {
    '^@internal/(.*)$': '<rootDir>/internal/$1',
    '^@mocks/(.*)$': '<rootDir>/__mocks__/$1'
  }
};
