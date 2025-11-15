Contexto Representa a los actores y sistemas externos definidos por la orquestación (docker-compose.yml (lines 1-182)) y la capa de observabilidad (observability/prometheus.yml (lines 1-17), observability/grafana-datasources.yml (lines 1-7)).

```mermaid
graph LR
  linkStyle default fill:#ffffff

  subgraph diagram ["System Context View: Interview Stack"]
    style diagram fill:#ffffff,stroke:#ffffff

    1["<div style='font-weight: bold'>Cliente API</div><div style='font-size: 70%; margin-top: 0px'>[Person]</div><div style='font-size: 80%; margin-top:10px'>Personas, pruebas<br />automatizadas o integraciones<br />que consumen el CRUD de<br />productos.</div>"]
    style 1 fill:#08427b,stroke:#052e56,color:#ffffff
    3["<div style='font-weight: bold'>Interview Stack</div><div style='font-size: 70%; margin-top: 0px'>[Software System]</div><div style='font-size: 80%; margin-top:10px'>Monorepo multi-runtime que<br />expone el mismo API.</div>"]
    style 3 fill:#ffffff,stroke:#444444,color:#444444
    4["<div style='font-weight: bold'>PostgreSQL Cluster</div><div style='font-size: 70%; margin-top: 0px'>[Software System]</div><div style='font-size: 80%; margin-top:10px'>Persistencia relacional<br />inicializada por<br />migrations/001_init.sql.</div>"]
    style 4 fill:#ffffff,stroke:#444444,color:#444444
    5["<div style='font-weight: bold'>MongoDB Replica</div><div style='font-size: 70%; margin-top: 0px'>[Software System]</div><div style='font-size: 80%; margin-top:10px'>Persistencia documental para<br />lecturas flexibles.</div>"]
    style 5 fill:#ffffff,stroke:#444444,color:#444444
    6["<div style='font-weight: bold'>Redis</div><div style='font-size: 70%; margin-top: 0px'>[Software System]</div><div style='font-size: 80%; margin-top:10px'>Cache distribuida y stream<br />para write-behind.</div>"]
    style 6 fill:#ffffff,stroke:#444444,color:#444444
    7["<div style='font-weight: bold'>Prometheus</div><div style='font-size: 70%; margin-top: 0px'>[Software System]</div><div style='font-size: 80%; margin-top:10px'>Scrapea métricas /metrics.</div>"]
    style 7 fill:#ffffff,stroke:#444444,color:#444444

    3-. "<div>CRUD transaccional<br />(database.type = sql)</div><div style='font-size: 70%'>[pgx/pg/Dapper]</div>" .->4
    3-. "<div>Persistencia documental<br />(database.type = mongo)</div><div style='font-size: 70%'>[Mongo drivers]</div>" .->5
    3-. "<div>Cachea y produce eventos</div><div style='font-size: 70%'>[Redis + XADD]</div>" .->6
    7-. "<div>Scrapea métricas</div><div style='font-size: 70%'>[HTTP pull]</div>" .->3
    1-. "<div>Gestiona productos</div><div style='font-size: 70%'>[HTTPS/JSON vía Nginx]</div>" .->3

  end
  ```
