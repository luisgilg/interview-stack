workspace "Interview Stack - Componentes Go" "Vista de componentes" {
  model {
    user = person "Cliente API" "Origen de las peticiones /go."
    system = softwareSystem "Interview Stack" {
      nginx = container "Nginx Edge Router" "Nginx" ""
      goApi = container "Go Products API" "Go + Fiber" "" {
        fiberRouter = component "Fiber Router" "fiber.App + middlewares" "Registra rutas y /metrics."
        controller = component "ProductController" "internal/interface/http" "Traduce HTTP a casos de uso."
        useCasesRead = component "UseCases (List/Get/Health)" "internal/application/usecase" "Consultas rápidas con caché."
        useCasesWrite = component "UseCases (Create/Update/Delete)" "internal/application/usecase" "Aplica validaciones, cacheWriter y write queue."
        productRepo = component "ProductRepository" "internal/infrastructure/repository" "Abstracción sobre ProductStore."
        sqlStore = component "PostgresProductStore" "internal/infrastructure/sql" "Opera via pgxpool."
        mongoStore = component "MongoProductStore" "internal/infrastructure/nosql" "Opera via mongo-driver."
        cacheService = component "CacheService" "internal/application/cache" "Lee/Escribe Redis."
        writeQueue = component "WriteQueueProducer" "internal/domain queue" "Encola eventos en Redis Stream."
        worker = component "WriteBehindWorker" "internal/infrastructure/queue" "Consume stream y persiste."
        metricsMw = component "PrometheusMiddleware" "internal/observability/metrics" "Expone http_requests_*."
      }
      redisCache = container "Redis" "Cache/Stream"
      postgresDb = container "PostgreSQL" ""
      mongoDb = container "MongoDB" ""
      prometheusSrv = container "Prometheus" ""
    }
    user -> nginx "HTTPS"
    nginx -> fiberRouter "Forward /go/*"
    fiberRouter -> metricsMw "Mide cada request"
    fiberRouter -> controller "Despacha ruta"
    controller -> useCasesRead "GET/health"
    controller -> useCasesWrite "POST/PUT/DELETE"
    useCasesRead -> productRepo "Consulta"
    useCasesRead -> cacheService "Hit/Stale fetch"
    useCasesWrite -> productRepo "Mutación directa (cuando no hay write-behind)"
    useCasesWrite -> cacheService "Upsert + invalidación"
    useCasesWrite -> writeQueue "Encola evento (write-behind)"
    productRepo -> sqlStore "SQL" 
    productRepo -> mongoStore "Mongo" 
    cacheService -> redisCache "GET/SET/EXPIRE"
    writeQueue -> redisCache "XADD stream"
    worker -> redisCache "XREADGROUP"
    worker -> productRepo "Aplica eventos"
    productRepo -> sqlStore
    productRepo -> mongoStore
    prometheusSrv -> metricsMw "Scrape"
  }
  views {
    component goApi componentes_go "Componentes Go" {
      include *
      autoLayout lr
    }
    styles {
      element "Component" {
        background "#85bbf0"
        color "#000000"
      }
      element "Container" {
        background "#438dd5"
        color "#ffffff"
      }
      element "Person" {
        background "#08427b"
        color "#ffffff"
      }
    }
  }
}
