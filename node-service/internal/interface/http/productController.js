const mapError = require('./errorHandler');

class RequestTimeoutError extends Error {
  constructor(operation, timeout) {
    super(`${operation} timed out after ${timeout}ms`);
    this.name = 'RequestTimeoutError';
    this.statusCode = 504;
  }
}

class ProductController {
  constructor(useCases, timeouts = {}) {
    this.listUC = useCases.list;
    this.getUC = useCases.get;
    this.createUC = useCases.create;
    this.updateUC = useCases.update;
    this.deleteUC = useCases.remove;
    this.healthUC = useCases.health;
    this.timeouts = {
      read: timeouts.read ?? 2000,
      write: timeouts.write ?? 5000,
      health: timeouts.health ?? 2000
    };
  }

  register(app) {
    /**
     * @openapi
     * /health:
     *   get:
     *     summary: Health check
     *     tags:
     *       - Health
     *     responses:
     *       200:
     *         description: Service is healthy
     *         content:
     *           application/json:
     *             schema:
     *               $ref: '#/components/schemas/HealthResponse'
     *       503:
     *         description: Service unavailable
     *         content:
     *           application/json:
     *             schema:
     *               $ref: '#/components/schemas/ErrorResponse'
     */
    app.get('/health', this.health.bind(this));

    /**
     * @openapi
     * /products:
     *   get:
     *     summary: List products
     *     tags:
     *       - Products
     *     responses:
     *       200:
     *         description: Products returned
     *         content:
     *           application/json:
     *             schema:
     *               type: array
     *               items:
     *                 $ref: '#/components/schemas/ProductResponse'
     *       500:
     *         description: Unexpected error
     *         content:
     *           application/json:
     *             schema:
     *               $ref: '#/components/schemas/ErrorResponse'
     */
    app.get('/products', this.list.bind(this));

    /**
     * @openapi
     * /products/{id}:
     *   get:
     *     summary: Get product
     *     tags:
     *       - Products
     *     parameters:
     *       - in: path
     *         name: id
     *         required: true
     *         schema:
     *           type: string
     *         description: Product identifier
     *     responses:
     *       200:
     *         description: Product returned
     *         content:
     *           application/json:
     *             schema:
     *               $ref: '#/components/schemas/ProductResponse'
     *       404:
     *         description: Product not found
     *         content:
     *           application/json:
     *             schema:
     *               $ref: '#/components/schemas/ErrorResponse'
     *       500:
     *         description: Unexpected error
     *         content:
     *           application/json:
     *             schema:
     *               $ref: '#/components/schemas/ErrorResponse'
     */
    app.get('/products/:id', this.get.bind(this));

    /**
     * @openapi
     * /products:
     *   post:
     *     summary: Create product
     *     tags:
     *       - Products
     *     requestBody:
     *       required: true
     *       content:
     *         application/json:
     *           schema:
     *             $ref: '#/components/schemas/ProductRequest'
     *           examples:
     *             default:
     *               value:
     *                 name: Premium Widget
     *                 price: 29.99
     *                 tags:
     *                   - gadget
     *                   - home
     *     responses:
     *       201:
     *         description: Product created
     *         content:
     *           application/json:
     *             schema:
     *               $ref: '#/components/schemas/ProductResponse'
     *       400:
     *         description: Validation failed
     *         content:
     *           application/json:
     *             schema:
     *               $ref: '#/components/schemas/ErrorResponse'
     *       500:
     *         description: Unexpected error
     *         content:
     *           application/json:
     *             schema:
     *               $ref: '#/components/schemas/ErrorResponse'
     */
    app.post('/products', this.create.bind(this));

    /**
     * @openapi
     * /products/{id}:
     *   put:
     *     summary: Update product
     *     tags:
     *       - Products
     *     parameters:
     *       - in: path
     *         name: id
     *         required: true
     *         schema:
     *           type: string
     *         description: Product identifier
     *     requestBody:
     *       required: true
     *       content:
     *         application/json:
     *           schema:
     *             $ref: '#/components/schemas/ProductRequest'
     *     responses:
     *       200:
     *         description: Product updated
     *         content:
     *           application/json:
     *             schema:
     *               $ref: '#/components/schemas/ProductResponse'
     *       400:
     *         description: Validation failed
     *         content:
     *           application/json:
     *             schema:
     *               $ref: '#/components/schemas/ErrorResponse'
     *       404:
     *         description: Product not found
     *         content:
     *           application/json:
     *             schema:
     *               $ref: '#/components/schemas/ErrorResponse'
     *       500:
     *         description: Unexpected error
     *         content:
     *           application/json:
     *             schema:
     *               $ref: '#/components/schemas/ErrorResponse'
     */
    app.put('/products/:id', this.update.bind(this));

    /**
     * @openapi
     * /products/{id}:
     *   delete:
     *     summary: Delete product
     *     tags:
     *       - Products
     *     parameters:
     *       - in: path
     *         name: id
     *         required: true
     *         schema:
     *           type: string
     *         description: Product identifier
     *     responses:
     *       204:
     *         description: Product deleted
     *       404:
     *         description: Product not found
     *         content:
     *           application/json:
     *             schema:
     *               $ref: '#/components/schemas/ErrorResponse'
     *       500:
     *         description: Unexpected error
     *         content:
     *           application/json:
     *             schema:
     *               $ref: '#/components/schemas/ErrorResponse'
     */
    app.delete('/products/:id', this.delete.bind(this));
  }

  async list(_req, res) {
    try {
      const products = await this.executeWithTimeout(
        () => this.listUC.execute(),
        this.timeouts.read,
        'list products'
      );
      res.json(products);
    } catch (err) {
      this.handleError(res, err);
    }
  }

  async get(req, res) {
    try {
      const product = await this.executeWithTimeout(
        () => this.getUC.execute(req.params.id),
        this.timeouts.read,
        'get product'
      );
      res.json(product);
    } catch (err) {
      this.handleError(res, err);
    }
  }

  async create(req, res) {
    try {
      const product = await this.executeWithTimeout(
        () => this.createUC.execute(req.body),
        this.timeouts.write,
        'create product'
      );
      res.status(201).json(product);
    } catch (err) {
      this.handleError(res, err);
    }
  }

  async update(req, res) {
    try {
      const product = await this.executeWithTimeout(
        () => this.updateUC.execute(req.params.id, req.body),
        this.timeouts.write,
        'update product'
      );
      res.json(product);
    } catch (err) {
      this.handleError(res, err);
    }
  }

  async delete(req, res) {
    try {
      await this.executeWithTimeout(
        () => this.deleteUC.execute(req.params.id),
        this.timeouts.write,
        'delete product'
      );
      res.status(204).send();
    } catch (err) {
      this.handleError(res, err);
    }
  }

  async health(_req, res) {
    try {
      await this.executeWithTimeout(() => this.healthUC.execute(), this.timeouts.health, 'health check');
      res.json({ status: 'ok' });
    } catch (err) {
      if (err instanceof RequestTimeoutError) {
        res.status(503).json({ status: 'unhealthy', error: err.message });
        return;
      }
      res.status(503).json({ status: 'unhealthy', error: err.message });
    }
  }

  async executeWithTimeout(work, timeout, operation) {
    let timer;
    try {
      return await Promise.race([
        work(),
        new Promise((_, reject) => {
          timer = setTimeout(() => reject(new RequestTimeoutError(operation, timeout)), timeout);
        })
      ]);
    } finally {
      clearTimeout(timer);
    }
  }

  handleError(res, err) {
    if (err instanceof RequestTimeoutError) {
      return res.status(504).json({ error: err.message });
    }
    return mapError(err, res);
  }
}

module.exports = ProductController;
