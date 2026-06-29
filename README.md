# Microservices - Order, Payment e Shipping

Sistema de microsserviços gRPC para gerenciamento de pedidos, pagamentos e entregas.

## Arquitetura

O projeto utiliza **Arquitetura Hexagonal** (Ports & Adapters) com comunicação via gRPC.

Cada serviço segue a estrutura:
- `cmd/` — ponto de entrada
- `config/` — leitura de variáveis de ambiente
- `internal/application/core/` — domínio e lógica de negócio
- `internal/ports/` — interfaces (ports)
- `internal/adapters/` — implementações (gRPC, MySQL)

## Pré-requisitos

- Docker e Docker Compose
- Go 1.21+ (para desenvolvimento local)
- grpcurl (opcional, para testes)

---

## Deploy com Docker

### 1. Iniciar todos os serviços

```powershell
cd microservices
docker compose up --build -d
```

Isso irá:
- Iniciar o MySQL com os bancos `order`, `payment` e `shipping`
- Construir e iniciar os três microsserviços

### 2. Inserir dados de estoque

```powershell
docker exec -i mysql mysql -uroot -pminhasenha order -e "
INSERT INTO stock_items (created_at, updated_at, deleted_at, product_code, name, unit_price) VALUES
  (NOW(), NOW(), NULL, 'PROD-001', 'Produto A', 10.00),
  (NOW(), NOW(), NULL, 'PROD-002', 'Produto B', 25.50),
  (NOW(), NOW(), NULL, 'PROD-003', 'Produto C', 5.99);"
```

### 3. Verificar status

```powershell
docker compose ps
```

### 4. Parar os serviços

```powershell
docker compose down
```

Para remover também os volumes (dados):

```powershell
docker compose down -v
```

---

## Desenvolvimento Local

### 1. Configurar dependências proto

Clone o repositório `microservices-proto` na mesma pasta pai que `microservices`:

```
pasta-pai/
├── microservices/
└── microservices-proto/
```

### 2. Iniciar MySQL

```powershell
docker run --name mysql-grpc -p 3306:3306 `
  -e MYSQL_ROOT_PASSWORD=minhasenha `
  -v "${PWD}\init.sql:/docker-entrypoint-initdb.d/init.sql" `
  -d mysql:8.0
```

### 3. Iniciar os serviços

**Terminal 1 - Payment:**
```powershell
cd payment
$env:DATA_SOURCE_URL="root:minhasenha@tcp(127.0.0.1:3306)/payment?charset=utf8mb4&parseTime=True&loc=Local"
$env:APPLICATION_PORT="3001"; $env:ENV="development"
go run cmd/main.go
```

**Terminal 2 - Shipping:**
```powershell
cd shipping
$env:DATA_SOURCE_URL="root:minhasenha@tcp(127.0.0.1:3306)/shipping?charset=utf8mb4&parseTime=True&loc=Local"
$env:APPLICATION_PORT="3002"; $env:ENV="development"
go run cmd/main.go
```

**Terminal 3 - Order:**
```powershell
cd order
$env:DATA_SOURCE_URL="root:minhasenha@tcp(127.0.0.1:3306)/order?charset=utf8mb4&parseTime=True&loc=Local"
$env:APPLICATION_PORT="3000"; $env:ENV="development"
$env:PAYMENT_SERVICE_URL="localhost:3001"
$env:SHIPPING_SERVICE_URL="localhost:3002"
go run cmd/main.go
```

### 4. Inserir dados de estoque

```powershell
docker exec -i mysql-grpc mysql -uroot -pminhasenha order -e "
INSERT INTO stock_items (created_at, updated_at, deleted_at, product_code, name, unit_price) VALUES
  (NOW(), NOW(), NULL, 'PROD-001', 'Produto A', 10.00),
  (NOW(), NOW(), NULL, 'PROD-002', 'Produto B', 25.50),
  (NOW(), NOW(), NULL, 'PROD-003', 'Produto C', 5.99);"
```

---

## Variáveis de Ambiente

| Variável | Serviço | Descrição |
|----------|---------|-----------|
| `DATA_SOURCE_URL` | todos | URL de conexão com MySQL |
| `APPLICATION_PORT` | todos | Porta do servidor gRPC |
| `ENV` | todos | `development` habilita gRPC reflection |
| `PAYMENT_SERVICE_URL` | order | Endereço do serviço Payment |
| `SHIPPING_SERVICE_URL` | order | Endereço do serviço Shipping |

---

## Testando com grpcurl

**Inserir dados de estoque (necessário antes de criar pedidos):**  
Ver seção "Inserir dados de estoque" acima.

**Criar pedido:**
```powershell
grpcurl -plaintext `
  -d "{\"costumer_id\":1,\"order_items\":[{\"product_code\":\"PROD-001\",\"quantity\":3,\"unit_price\":10}]}" `
  localhost:3000 Order/Create
```

**Testar Payment diretamente:**
```powershell
grpcurl -plaintext `
  -d "{\"user_id\":1,\"order_id\":1,\"total_price\":100}" `
  localhost:3001 Payment/Create
```

**Testar Shipping diretamente:**
```powershell
grpcurl -plaintext `
  -d "{\"order_id\":1,\"items\":[{\"product_code\":\"PROD-001\",\"quantity\":5}]}" `
  localhost:3002 Shipping/Create
```

> Sem grpcurl instalado, use a imagem Docker:
> ```powershell
> $json = '{"costumer_id":1,"order_items":[{"product_code":"PROD-001","quantity":3,"unit_price":10}]}'
> $json | docker run -i --rm fullstorydev/grpcurl:latest -plaintext -d '@' host.docker.internal:3000 Order/Create
> ```

---

## Validações

| Regra | Código gRPC | Mensagem |
|-------|-------------|----------|
| Produto não existe no estoque | `NOT_FOUND` | `product_code 'X' does not exist in stock` |
| Quantidade total > 50 itens | `INVALID_ARGUMENT` | `Order cannot have more than 50 items in total.` |
| Total do pagamento > R$1.000,00 | `INVALID_ARGUMENT` | `Payment over 1000 is not allowed.` |
| Erro no banco de dados | `INTERNAL` | — |
| Timeout na chamada ao Payment ou Shipping (2s) | `DEADLINE_EXCEEDED` | registrado em log |
| Serviço indisponível | `UNAVAILABLE` | retry automático até 5x |

---

## Cálculo de Dias de Entrega

O prazo é calculado pelo serviço Shipping com base no total de unidades:

```
dias = 1 + (total_unidades / 5)
```

| Unidades | Prazo |
|----------|-------|
| 1 – 4    | 1 dia |
| 5 – 9    | 2 dias |
| 10 – 14  | 3 dias |

---

## Estrutura do Projeto

```
microservices/
├── docker-compose.yml
├── init.sql                  # Cria os bancos order, payment e shipping
├── k8s/                      # Manifests Kubernetes
├── order/                    # Microsserviço Order   (porta 3000)
├── payment/                  # Microsserviço Payment (porta 3001)
└── shipping/                 # Microsserviço Shipping (porta 3002)
```

---

## Licença

Projeto acadêmico — IFPB
