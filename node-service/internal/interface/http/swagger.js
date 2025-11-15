const path = require('path');
const swaggerJSDoc = require('swagger-jsdoc');

const swaggerDefinition = {
  openapi: '3.0.0',
  info: {
    title: 'Products API (Node)',
    version: '1.0.0',
    description: 'REST API powered by Express for managing products.'
  },
  servers: [
    {
      url: 'http://localhost:8082',
      description: 'Local Node.js service'
    }
  ],
  components: {
    schemas: {
      ProductRequest: {
        type: 'object',
        required: ['name', 'price'],
        properties: {
          name: {
            type: 'string',
            example: 'Premium Widget'
          },
          price: {
            type: 'number',
            format: 'float',
            example: 29.99
          },
          tags: {
            type: 'array',
            items: { type: 'string' },
            example: ['gadget', 'home']
          }
        }
      },
      ProductResponse: {
        allOf: [
          {
            type: 'object',
            properties: {
              id: {
                type: 'string',
                example: '507f1f77bcf86cd799439011'
              },
              createdAt: {
                type: 'string',
                format: 'date-time',
                example: '2024-01-01T10:00:00Z'
              },
              updatedAt: {
                type: 'string',
                format: 'date-time',
                example: '2024-01-01T11:00:00Z'
              },
              cache_status: {
                type: 'string',
                example: 'fresh'
              }
            }
          },
          { $ref: '#/components/schemas/ProductRequest' }
        ]
      },
      ErrorResponse: {
        type: 'object',
        properties: {
          error: {
            type: 'string',
            example: 'product not found'
          }
        }
      },
      HealthResponse: {
        type: 'object',
        properties: {
          status: {
            type: 'string',
            example: 'ok'
          }
        }
      }
    }
  }
};

const swaggerSpec = swaggerJSDoc({
  definition: swaggerDefinition,
  apis: [path.join(__dirname, 'productController.js')]
});

module.exports = { swaggerSpec };
