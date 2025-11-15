Contenedores Desglosa la plataforma en edge, servicios, datos y observabilidad según el router, los bootsraps de cada servicio y los recursos declarados en Compose.

```mermaid
graph LR
  linkStyle default fill:#ffffff

  subgraph diagram ["Container View: Interview Stack"]
    style diagram fill:#ffffff,stroke:#ffffff

    1["<div style='font-weight: bold'>Cliente API</div><div style='font-size: 70%; margin-top: 0px'>[Person]</div><div style='font-size: 80%; margin-top:10px'>Usa /go, /node o /dotnet.</div>"]
    style 1 fill:#08427b,stroke:#052e56,color:#ffffff

    subgraph 2 ["Interview Stack"]
      style 2 fill:#ffffff,stroke:#444444,color:#444444

      10["<div style='font-weight: bold'>Prometheus 2.53</div><div style='font-size: 70%; margin-top: 0px'>[Container]</div><div style='font-size: 80%; margin-top:10px'>Scrapea /metrics</div>"]
      style 10 fill:#438dd5,stroke:#2e6295,color:#ffffff
      11["<div style='font-weight: bold'>Grafana 11.2</div><div style='font-size: 70%; margin-top: 0px'>[Container]</div><div style='font-size: 80%; margin-top:10px'>Dashboards provisionados</div>"]
      style 11 fill:#438dd5,stroke:#2e6295,color:#ffffff
      3["<div style='font-weight: bold'>Nginx Edge Router</div><div style='font-size: 70%; margin-top: 0px'>[Container: Publica /go/, /node/ y /dotnet/.]</div><div style='font-size: 80%; margin-top:10px'>Nginx 1.25</div>"]
      style 3 fill:#438dd5,stroke:#2e6295,color:#ffffff
      4["<div style='font-weight: bold'>Go Products API</div><div style='font-size: 70%; margin-top: 0px'>[Container: CRUD + métricas.]</div><div style='font-size: 80%; margin-top:10px'>Go 1.22 + Fiber</div>"]
      style 4 fill:#438dd5,stroke:#2e6295,color:#ffffff
      5["<div style='font-weight: bold'>Node Products API</div><div style='font-size: 70%; margin-top: 0px'>[Container: CRUD + métricas.]</div><div style='font-size: 80%; margin-top:10px'>Node 20 + Express</div>"]
      style 5 fill:#438dd5,stroke:#2e6295,color:#ffffff
      6["<div style='font-weight: bold'>.NET Products API</div><div style='font-size: 70%; margin-top: 0px'>[Container: CRUD + métricas.]</div><div style='font-size: 80%; margin-top:10px'>.NET 8 Minimal API</div>"]
      style 6 fill:#438dd5,stroke:#2e6295,color:#ffffff
      7["<div style='font-weight: bold'>PostgreSQL 16</div><div style='font-size: 70%; margin-top: 0px'>[Container: productsdb]</div><div style='font-size: 80%; margin-top:10px'>Base relacional</div>"]
      style 7 fill:#438dd5,stroke:#2e6295,color:#ffffff
      8["<div style='font-weight: bold'>MongoDB 7</div><div style='font-size: 70%; margin-top: 0px'>[Container]</div><div style='font-size: 80%; margin-top:10px'>Colección products</div>"]
      style 8 fill:#438dd5,stroke:#2e6295,color:#ffffff
      9["<div style='font-weight: bold'>Redis 7.4</div><div style='font-size: 70%; margin-top: 0px'>[Container]</div><div style='font-size: 80%; margin-top:10px'>Cache TTL + stream<br />products_write_queue</div>"]
      style 9 fill:#438dd5,stroke:#2e6295,color:#ffffff
    end

    1-. "<div>HTTPS/JSON</div><div style='font-size: 70%'></div>" .->3
    3-. "<div>/go/*</div><div style='font-size: 70%'></div>" .->4
    3-. "<div>/node/*</div><div style='font-size: 70%'></div>" .->5
    3-. "<div>/dotnet/*</div><div style='font-size: 70%'></div>" .->6
    4-. "<div>SQL (database.type = sql)</div><div style='font-size: 70%'></div>" .->7
    4-. "<div>Mongo (database.type = mongo)</div><div style='font-size: 70%'></div>" .->8
    4-. "<div>Cache y stream</div><div style='font-size: 70%'></div>" .->9
    5-. "<div>SQL (database.type = sql)</div><div style='font-size: 70%'></div>" .->7
    5-. "<div>Mongo (database.type = mongo)</div><div style='font-size: 70%'></div>" .->8
    5-. "<div></div><div style='font-size: 70%'></div>" .->9
    6-. "<div>SQL (database.type = sql)</div><div style='font-size: 70%'></div>" .->7
    6-. "<div>Mongo (database.type = mongo)</div><div style='font-size: 70%'></div>" .->8
    6-. "<div></div><div style='font-size: 70%'></div>" .->9
    10-. "<div>GET /metrics</div><div style='font-size: 70%'></div>" .->4
    10-. "<div></div><div style='font-size: 70%'></div>" .->5
    10-. "<div></div><div style='font-size: 70%'></div>" .->6
    11-. "<div>PromQL</div><div style='font-size: 70%'></div>" .->10

  end
  ```
