# microservices

Projeto da disciplina **Programação Distribuída** (Prática - Microsserviços com gRPC).

Este repositório contém a implementação do microsserviço **Order** em Go, seguindo arquitetura hexagonal (ports and adapters), com:
- API gRPC
- persistência em MySQL (GORM)
- integração com stubs protobuf do repositório `microservices-proto`

## Estrutura

- `order/cmd/main.go`: bootstrap do serviço
- `order/config`: leitura de variáveis de ambiente
- `order/internal/application/core`: domínio e lógica de aplicação
- `order/internal/ports`: interfaces (ports)
- `order/internal/adapters`: adaptadores (gRPC e MySQL)

## Como executar

Pré-requisitos:
- Go instalado
- Docker Desktop em execução

### 1) Subir MySQL com Docker

```bash
docker run --name mysql-order -p 3306:3306 -e MYSQL_ROOT_PASSWORD=minhasenha -e MYSQL_DATABASE=order -d mysql:8
```

Se a porta `3306` estiver ocupada, use:

```bash
docker run --name mysql-order -p 3307:3306 -e MYSQL_ROOT_PASSWORD=minhasenha -e MYSQL_DATABASE=order -d mysql:8
```

### 2) Rodar o serviço Order

Na pasta `order`:

```powershell
cd order
$env:DATA_SOURCE_URL="root:minhasenha@tcp(127.0.0.1:3306)/order"
$env:APPLICATION_PORT="3000"
$env:ENV="development"
go run cmd/main.go
```

Se estiver usando MySQL na porta `3307`, ajuste a URL:

```powershell
$env:DATA_SOURCE_URL="root:minhasenha@tcp(127.0.0.1:3307)/order"
```

### 3) Testar chamada gRPC

Com grpcurl instalado localmente:

```powershell
grpcurl -plaintext -d "{\"costumer_id\":123,\"order_items\":[{\"product_code\":\"prod\",\"quantity\":4,\"unit_price\":12}]}" localhost:3000 Order/Create
```

Alternativa sem instalar grpcurl (usando Docker):

```powershell
$json = '{"costumer_id":123,"order_items":[{"product_code":"prod","quantity":4,"unit_price":12}]}'
$json | docker run -i --rm fullstorydev/grpcurl:latest -plaintext -d '@' host.docker.internal:3000 Order/Create
```

## Resultado esperado no teste

Resposta gRPC com `orderId` e registro persistido no banco.

Exemplo de validação no MySQL:

```bash
docker exec mysql-order mysql -uroot -pminhasenha -e "use order; select id, customer_id, status from orders order by id desc limit 5;"
```
