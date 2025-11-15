workspace "Interview Stack - Contenedores" "Vista de contenedores" {
  model {
    user = person "Cliente API" "Usa /go, /node o /dotnet."
    system = softwareSystem "Interview Stack" "Plataforma multi-runtime" {
      nginx = container "Nginx Edge Router" "Nginx 1.25" "Publica /go/, /node/ y /dotnet/."
      goApi = container "Go Products API" "Go 1.22 + Fiber" "CRUD + métricas."
      nodeApi = container "Node Products API" "Node 20 + Express" "CRUD + métricas."
      dotnetApi = container ".NET Products API" ".NET 8 Minimal API" "CRUD + métricas."
      postgresDb = container "PostgreSQL 16" "Base relacional" "productsdb"
      mongoDb = container "MongoDB 7" "Colección products"
      redisCache = container "Redis 7.4" "Cache TTL + stream products_write_queue"
      prometheusSrv = container "Prometheus 2.53" "Scrapea /metrics"
      grafanaSrv = container "Grafana 11.2" "Dashboards provisionados"
    }
    user -> nginx "HTTPS/JSON"
    nginx -> goApi "/go/*"
    nginx -> nodeApi "/node/*"
    nginx -> dotnetApi "/dotnet/*"
    goApi -> postgresDb "SQL (configurable)"
    goApi -> mongoDb "Mongo (configurable)"
    goApi -> redisCache "Cache y stream"
    nodeApi -> postgresDb
    nodeApi -> mongoDb
    nodeApi -> redisCache
    dotnetApi -> postgresDb
    dotnetApi -> mongoDb
    dotnetApi -> redisCache
    prometheusSrv -> goApi "GET /metrics"
    prometheusSrv -> nodeApi
    prometheusSrv -> dotnetApi
    grafanaSrv -> prometheusSrv "PromQL"
  }
  views {
    container system contenedores "Mapa de contenedores" {
      include *
      autoLayout lr
    }
    styles {
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
