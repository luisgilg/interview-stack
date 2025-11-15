# Sample cURL Commands

Use these snippets to interact with the Product APIs exposed by the three services. Replace the port (`8081/8082/8083`) with the service you are targeting (Go, Node, or .NET).

## Health Check

```bash
curl -i http://localhost:8081/health
```

## List Products

```bash
curl -i http://localhost:8081/products
```

## Create Product

```bash
curl -i -X POST http://localhost:8081/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Mechanical Keyboard",
    "price": 129.99,
    "tags": ["peripherals", "gaming"]
  }'
```

## Get Product by ID

```bash
curl -i http://localhost:8081/products/<PRODUCT_ID>
```

## Update Product

```bash
curl -i -X PUT http://localhost:8081/products/<PRODUCT_ID> \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Wireless Keyboard",
    "price": 149.99,
    "tags": ["peripherals", "wireless"]
  }'
```

## Delete Product

```bash
curl -i -X DELETE http://localhost:8081/products/<PRODUCT_ID>
```

## Switching Services

Run the same commands against:

- Go service: `http://localhost:8081`
- Node service: `http://localhost:8082`
- .NET service: `http://localhost:8083`

Everything else (payloads, headers) remains identical across services.
