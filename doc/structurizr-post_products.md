Flujo POST /products (Nivel 4) El diagrama dinámico resume la ruta de escritura, el cacheWriter y el trabajador Redis descritos en go-service/internal/application/usecase/create_product.go (lines 24-74), go-service/internal/infrastructure/queue/worker.go (lines 22-199) y activados desde go-service/cmd/http/main.go (lines 121-187). El mismo patrón se replica en Node y .NET (node-service/cmd/http/server.js (lines 79-152), dotnet-service/src/cmd/Http/Program.cs (lines 79-177)).

```mermaid
graph LR
  linkStyle default fill:#ffffff

  subgraph diagram ["Dynamic View: Interview Stack - Go Products API"]
    style diagram fill:#ffffff,stroke:#ffffff

    1["<div style='font-weight: bold'>Cliente API</div><div style='font-size: 70%; margin-top: 0px'>[Person]</div>"]
    style 1 fill:#08427b,stroke:#052e56,color:#ffffff

    subgraph 2 ["Interview Stack"]
      style 2 fill:#ffffff,stroke:#444444,color:#444444

      subgraph 4 ["Go Products API"]
        style 4 fill:#ffffff,stroke:#2e6295,color:#2e6295

        10["<div style='font-weight: bold'>WriteQueueProducer</div><div style='font-size: 70%; margin-top: 0px'>[Component]</div>"]
        style 10 fill:#85bbf0,stroke:#5d82a8,color:#000000
        11["<div style='font-weight: bold'>WriteBehindWorker</div><div style='font-size: 70%; margin-top: 0px'>[Component]</div>"]
        style 11 fill:#85bbf0,stroke:#5d82a8,color:#000000
        12["<div style='font-weight: bold'>PostgresProductStore</div><div style='font-size: 70%; margin-top: 0px'>[Component]</div>"]
        style 12 fill:#85bbf0,stroke:#5d82a8,color:#000000
        13["<div style='font-weight: bold'>MongoProductStore</div><div style='font-size: 70%; margin-top: 0px'>[Component]</div>"]
        style 13 fill:#85bbf0,stroke:#5d82a8,color:#000000
        5["<div style='font-weight: bold'>Fiber Router</div><div style='font-size: 70%; margin-top: 0px'>[Component]</div>"]
        style 5 fill:#85bbf0,stroke:#5d82a8,color:#000000
        6["<div style='font-weight: bold'>ProductController</div><div style='font-size: 70%; margin-top: 0px'>[Component]</div>"]
        style 6 fill:#85bbf0,stroke:#5d82a8,color:#000000
        7["<div style='font-weight: bold'>CreateProductUseCase</div><div style='font-size: 70%; margin-top: 0px'>[Component]</div>"]
        style 7 fill:#85bbf0,stroke:#5d82a8,color:#000000
        8["<div style='font-weight: bold'>ProductRepository</div><div style='font-size: 70%; margin-top: 0px'>[Component]</div>"]
        style 8 fill:#85bbf0,stroke:#5d82a8,color:#000000
        9["<div style='font-weight: bold'>CacheService</div><div style='font-size: 70%; margin-top: 0px'>[Component]</div>"]
        style 9 fill:#85bbf0,stroke:#5d82a8,color:#000000
      end

      14["<div style='font-weight: bold'>Redis</div><div style='font-size: 70%; margin-top: 0px'>[Container]</div>"]
      style 14 fill:#438dd5,stroke:#2e6295,color:#ffffff
      3["<div style='font-weight: bold'>Nginx Edge</div><div style='font-size: 70%; margin-top: 0px'>[Container]</div>"]
      style 3 fill:#438dd5,stroke:#2e6295,color:#ffffff
    end

    1["<div style='font-weight: bold'>Cliente API</div><div style='font-size: 70%; margin-top: 0px'>[Person]</div>"]
    style 1 fill:#08427b,stroke:#052e56,color:#ffffff
    3["<div style='font-weight: bold'>Nginx Edge</div><div style='font-size: 70%; margin-top: 0px'>[Container]</div>"]
    style 3 fill:#438dd5,stroke:#2e6295,color:#ffffff
    14["<div style='font-weight: bold'>Redis</div><div style='font-size: 70%; margin-top: 0px'>[Container]</div>"]
    style 14 fill:#438dd5,stroke:#2e6295,color:#ffffff

    1-. "<div>1. 1. Solicita POST /products</div><div style='font-size: 70%'></div>" .->3
    3-. "<div>2. 2. Proxy de Nginx a Fiber</div><div style='font-size: 70%'></div>" .->5
    5-. "<div>3. 3. Router invoca<br />ProductController</div><div style='font-size: 70%'></div>" .->6
    6-. "<div>4. 4. Caso de uso procesa</div><div style='font-size: 70%'></div>" .->7
    7-. "<div>5. 5. Persistencia directa<br />opcional</div><div style='font-size: 70%'></div>" .->8
    8-. "<div>6. 6. Persiste en PostgreSQL<br />(database.type = sql)</div><div style='font-size: 70%'></div>" .->12
    8-. "<div>7. 7. Persiste en MongoDB<br />(database.type = mongo)</div><div style='font-size: 70%'></div>" .->13
    7-. "<div>8. 8. Actualiza caché/lista</div><div style='font-size: 70%'></div>" .->9
    9-. "<div>9. 9. SET/DEL en Redis</div><div style='font-size: 70%'></div>" .->14
    7-. "<div>10. 10. Encola evento<br />write-behind</div><div style='font-size: 70%'></div>" .->10
    10-. "<div>11. 11. XADD stream</div><div style='font-size: 70%'></div>" .->14
    11-. "<div>12. 12. Consume stream</div><div style='font-size: 70%'></div>" .->14
    11-. "<div>13. 13. Aplica evento</div><div style='font-size: 70%'></div>" .->8
    6-. "<div>14. 14. Devuelve DTO</div><div style='font-size: 70%'></div>" .->5
    5-. "<div>15. 15. Respuesta HTTP 201</div><div style='font-size: 70%'></div>" .->1

  end
  ```
