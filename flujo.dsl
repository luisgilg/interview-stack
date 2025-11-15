workspace "Interview Stack - Flujo POST /products" "Vista dinámica" {
  model {
    user = person "Cliente API" ""
    system = softwareSystem "Interview Stack" {
      nginx = container "Nginx Edge" ""
      goApi = container "Go Products API" "" {
        fiberRouter = component "Fiber Router" ""
        controller = component "ProductController" ""
        createUseCase = component "CreateProductUseCase" ""
        productRepo = component "ProductRepository" ""
        cacheService = component "CacheService" ""
        writeQueue = component "WriteQueueProducer" ""
        worker = component "WriteBehindWorker" ""
        sqlStore = component "PostgresProductStore" ""
        mongoStore = component "MongoProductStore" ""
      }
      redisCache = container "Redis" ""
    }
    user -> nginx
    nginx -> fiberRouter
    fiberRouter -> controller
    controller -> createUseCase
    createUseCase -> productRepo "Persistencia directa (writeBehind=false)"
    productRepo -> sqlStore
    productRepo -> mongoStore
    createUseCase -> cacheService "Actualiza cache/lista"
    cacheService -> redisCache "SET/DEL"
    createUseCase -> writeQueue "Encola evento (writeBehind=true)"
    writeQueue -> redisCache "XADD stream"
    worker -> redisCache "XREADGROUP"
    worker -> productRepo "Aplica evento"
    controller -> fiberRouter "DTO -> HTTP 201"
    fiberRouter -> user "Respuesta"
  }
  views {
    dynamic goApi post_products "Secuencia POST /products" {
      autoLayout lr
      user -> nginx "1. Solicita POST /products"
      nginx -> fiberRouter "2. Proxy de Nginx a Fiber"
      fiberRouter -> controller "3. Router invoca ProductController"
      controller -> createUseCase "4. Caso de uso procesa"
      createUseCase -> productRepo "5. Persistencia directa opcional"
      productRepo -> sqlStore "6. Escribe en PostgreSQL"
      productRepo -> mongoStore "7. Espeja en Mongo"
      createUseCase -> cacheService "8. Actualiza caché/lista"
      cacheService -> redisCache "9. SET/DEL en Redis"
      createUseCase -> writeQueue "10. Encola evento write-behind"
      writeQueue -> redisCache "11. XADD stream"
      worker -> redisCache "12. Consume stream"
      worker -> productRepo "13. Aplica evento"
      controller -> fiberRouter "14. Devuelve DTO"
      fiberRouter -> user "15. Respuesta HTTP 201"
    }
    styles {
      element "Person" {
        background "#08427b"
        color "#ffffff"
      }
      element "Component" {
        background "#85bbf0"
        color "#000000"
      }
      element "Container" {
        background "#438dd5"
        color "#ffffff"
      }
    }
  }
}
