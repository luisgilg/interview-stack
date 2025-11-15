Componentes (Go) Profundiza en la porción Go: el router Fiber y el controlador invocan los casos de uso y coordinan caché y write-behind que se configuran al arrancar.

```mermaid
graph LR
  linkStyle default fill:#ffffff

  subgraph diagram ["Component View: Interview Stack - Go Products API"]
    style diagram fill:#ffffff,stroke:#ffffff

    subgraph 2 ["Interview Stack"]
      style 2 fill:#ffffff,stroke:#444444,color:#444444

      subgraph 4 ["Go Products API"]
        style 4 fill:#ffffff,stroke:#2e6295,color:#2e6295

        10["<div style='font-weight: bold'>PostgresProductStore</div><div style='font-size: 70%; margin-top: 0px'>[Component: Opera via pgxpool.]</div><div style='font-size: 80%; margin-top:10px'>internal/infrastructure/sql</div>"]
        style 10 fill:#85bbf0,stroke:#5d82a8,color:#000000
        11["<div style='font-weight: bold'>MongoProductStore</div><div style='font-size: 70%; margin-top: 0px'>[Component: Opera via mongo-driver.]</div><div style='font-size: 80%; margin-top:10px'>internal/infrastructure/nosql</div>"]
        style 11 fill:#85bbf0,stroke:#5d82a8,color:#000000
        12["<div style='font-weight: bold'>CacheService</div><div style='font-size: 70%; margin-top: 0px'>[Component: Lee/Escribe Redis.]</div><div style='font-size: 80%; margin-top:10px'>internal/application/cache</div>"]
        style 12 fill:#85bbf0,stroke:#5d82a8,color:#000000
        13["<div style='font-weight: bold'>WriteQueueProducer</div><div style='font-size: 70%; margin-top: 0px'>[Component: Encola eventos en Redis Stream.]</div><div style='font-size: 80%; margin-top:10px'>internal/domain queue</div>"]
        style 13 fill:#85bbf0,stroke:#5d82a8,color:#000000
        14["<div style='font-weight: bold'>WriteBehindWorker</div><div style='font-size: 70%; margin-top: 0px'>[Component: Consume stream y persiste.]</div><div style='font-size: 80%; margin-top:10px'>internal/infrastructure/queue</div>"]
        style 14 fill:#85bbf0,stroke:#5d82a8,color:#000000
        15["<div style='font-weight: bold'>PrometheusMiddleware</div><div style='font-size: 70%; margin-top: 0px'>[Component: Expone http_requests_*.]</div><div style='font-size: 80%; margin-top:10px'>internal/observability/metrics</div>"]
        style 15 fill:#85bbf0,stroke:#5d82a8,color:#000000
        5["<div style='font-weight: bold'>Fiber Router</div><div style='font-size: 70%; margin-top: 0px'>[Component: Registra rutas y /metrics.]</div><div style='font-size: 80%; margin-top:10px'>fiber.App + middlewares</div>"]
        style 5 fill:#85bbf0,stroke:#5d82a8,color:#000000
        6["<div style='font-weight: bold'>ProductController</div><div style='font-size: 70%; margin-top: 0px'>[Component: Traduce HTTP a casos de uso.]</div><div style='font-size: 80%; margin-top:10px'>internal/interface/http</div>"]
        style 6 fill:#85bbf0,stroke:#5d82a8,color:#000000
        7["<div style='font-weight: bold'>UseCases (List/Get/Health)</div><div style='font-size: 70%; margin-top: 0px'>[Component: Consultas rápidas con caché.]</div><div style='font-size: 80%; margin-top:10px'>internal/application/usecase</div>"]
        style 7 fill:#85bbf0,stroke:#5d82a8,color:#000000
        8["<div style='font-weight: bold'>UseCases (Create/Update/Delete)</div><div style='font-size: 70%; margin-top: 0px'>[Component: Aplica validaciones, cacheWriter y write queue.]</div><div style='font-size: 80%; margin-top:10px'>internal/application/usecase</div>"]
        style 8 fill:#85bbf0,stroke:#5d82a8,color:#000000
        9["<div style='font-weight: bold'>ProductRepository</div><div style='font-size: 70%; margin-top: 0px'>[Component: Abstracción sobre ProductStore.]</div><div style='font-size: 80%; margin-top:10px'>internal/infrastructure/repository</div>"]
        style 9 fill:#85bbf0,stroke:#5d82a8,color:#000000
      end

      16["<div style='font-weight: bold'>Redis</div><div style='font-size: 70%; margin-top: 0px'>[Container]</div><div style='font-size: 80%; margin-top:10px'>Cache/Stream</div>"]
      style 16 fill:#438dd5,stroke:#2e6295,color:#ffffff
      19["<div style='font-weight: bold'>Prometheus</div><div style='font-size: 70%; margin-top: 0px'>[Container]</div>"]
      style 19 fill:#438dd5,stroke:#2e6295,color:#ffffff
      3["<div style='font-weight: bold'>Nginx Edge Router</div><div style='font-size: 70%; margin-top: 0px'>[Container]</div><div style='font-size: 80%; margin-top:10px'>Nginx</div>"]
      style 3 fill:#438dd5,stroke:#2e6295,color:#ffffff
    end

    3-. "<div>Forward /go/*</div><div style='font-size: 70%'></div>" .->5
    5-. "<div>Mide cada request</div><div style='font-size: 70%'></div>" .->15
    5-. "<div>Despacha ruta</div><div style='font-size: 70%'></div>" .->6
    6-. "<div>GET/health</div><div style='font-size: 70%'></div>" .->7
    6-. "<div>POST/PUT/DELETE</div><div style='font-size: 70%'></div>" .->8
    7-. "<div>Consulta</div><div style='font-size: 70%'></div>" .->9
    7-. "<div>Hit/Stale fetch</div><div style='font-size: 70%'></div>" .->12
    8-. "<div>Mutación directa (cuando no<br />hay write-behind)</div><div style='font-size: 70%'></div>" .->9
    8-. "<div>Upsert + invalidación</div><div style='font-size: 70%'></div>" .->12
    8-. "<div>Encola evento (write-behind)</div><div style='font-size: 70%'></div>" .->13
    9-. "<div>SQL (database.type = sql)</div><div style='font-size: 70%'></div>" .->10
    9-. "<div>Mongo (database.type = mongo)</div><div style='font-size: 70%'></div>" .->11
    12-. "<div>GET/SET/EXPIRE</div><div style='font-size: 70%'></div>" .->16
    13-. "<div>XADD stream</div><div style='font-size: 70%'></div>" .->16
    14-. "<div>XREADGROUP</div><div style='font-size: 70%'></div>" .->16
    14-. "<div>Aplica eventos</div><div style='font-size: 70%'></div>" .->9
    9-. "<div></div><div style='font-size: 70%'></div>" .->10
    9-. "<div></div><div style='font-size: 70%'></div>" .->11
    19-. "<div>Scrape</div><div style='font-size: 70%'></div>" .->15

  end
  ```
