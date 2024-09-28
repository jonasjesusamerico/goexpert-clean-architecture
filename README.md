# Desafio: Clean Architecture

## Executando o Projeto

1. Clone o repositório para o seu ambiente local:
   ```sh
   git clone https://github.com/jonasjesusamerico/goexpert-clean-architecture.git
    ```
2. Abra um terminal na raiz do projeto.
3. Execute o seguinte comando para construir e iniciar os serviços definidos no arquivo docker-compose:
   ```sh
   docker-compose build;
   ```
4. Execute o seguinte comando para iniciar a aplicação
   ```sh
   docker-compose up -d;
   ```

### API HTTP/Web

- **Para criar um novo pedido:**

  ```sh
  curl --location 'http://localhost:8000/order' \
  --header 'Content-Type: application/json' \
  --data '{
      "id": "1",
      "price": 10.0,
      "tax": 0.10
  }'
  ```

- **Para listar todos os pedidos criado:**

  ```sh
  curl --location 'http://localhost:8000/order'
  ```

### GraphQL

- **Para criar um novo pedido:**

  ```graphql
  mutation createOrder {
    createOrder(input: {id: "2", Price: 20.0, Tax: 0.20}) {
      id
      Price
      Tax
      FinalPrice
    }
  }
  ```

- **Para listar todos os pedidos criado:**

  ```graphql
  query  {
    listOrders {
      id
      Price
      Tax
      FinalPrice
    }
  }
  ```

### GRPC

**Para criar um novo pedido:**

      grpcurl -plaintext -d '{"id":"3","price": 30, "tax": 0.30}' localhost:50051 pb.OrderService/CreateOrder

**Para listar todos os pedidos criado:**

    grpcurl -plaintext -d '{}' localhost:50051 pb.OrderService/ListOrders

## PÓS EXECUÇÃO DOS TESTES

1. Execute o seguinte comando para remover todos os container criado
   ```sh
   docker-compose down -d;
   ```

### Observação

- Talvez se faça necessário o uso do super usuario no linux