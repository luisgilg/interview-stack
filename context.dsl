workspace "Interview Stack - Contexto" "Vista de contexto" {
  model {
    user = person "Cliente API" "Personas, pruebas automatizadas o integraciones que consumen el CRUD de productos."
    ops = person "Equipo de Observabilidad" "Monitoriza la salud y los KPIs."
    system = softwareSystem "Interview Stack" "Monorepo multi-runtime que expone el mismo API."
    postgres = softwareSystem "PostgreSQL Cluster" "Persistencia relacional inicializada por migrations/001_init.sql."
    mongo = softwareSystem "MongoDB Replica" "Persistencia documental para lecturas flexibles."
    redis = softwareSystem "Redis" "Cache distribuida y stream para write-behind."
    prometheus = softwareSystem "Prometheus" "Scrapea métricas /metrics."
    grafana = softwareSystem "Grafana" "Dashboards provisionados en observability/grafana-dashboard.json."
    user -> system "Gestiona productos" "HTTPS/JSON vía Nginx"
    system -> postgres "CRUD transaccional" "pgx/pg/Dapper"
    system -> mongo "Espeja documentos" "Mongo drivers"
    system -> redis "Cachea y produce eventos" "Redis + XADD"
    prometheus -> system "Scrapea métricas" "HTTP pull"
    grafana -> prometheus "Consulta series" "PromQL"
    ops -> grafana "Visualiza KPIs" "HTTP/Web"
  }
  views {
    systemContext system contexto "Vista general del sistema" {
      include *
      autoLayout lr
    }
    styles {
      element "Person" {
        background "#08427b"
        color "#ffffff"
      }
      element "SoftwareSystem" {
        background "#1168bd"
        color "#ffffff"
      }
    }
  }
}
