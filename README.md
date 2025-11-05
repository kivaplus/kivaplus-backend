# 🏢 KivaPlus Backend

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker)](https://www.docker.com/)

Sistema de gestão condominial construído com arquitetura serverless usando AWS Lambda, API Gateway, PostgreSQL, Redis e autenticação JWT.

## 📋 Pré-requisitos

- **Docker** (versão 20.10+)
- **Docker Compose** (versão 2.0+)
- **Go** (versão 1.21+)
- **Git**
- **curl** e **jq** (para testes)

## 🚀 Setup Rápido (5 minutos)

### 1. Clone o repositório
```bash
git clone https://github.com/kivaplus/kivaplus-backend.git
cd kivaplus-backend
```

### 2. Execute o setup completo
```bash
make setup
```

### 3. Verifique se tudo está funcionando
```bash
make test-integration
```

### 4. Verifique o status do sistema
```bash
make docker-status
```

## 🎯 Comandos Essenciais

| Comando | Descrição |
|---------|-----------|
| `make setup` | Setup completo do ambiente |
| `make build-all` | Construir todas as funções Lambda |
| `make test` | Executar testes unitários |
| `make test-integration` | Executar testes de integração |
| `make test-all` | Executar todos os testes |
| `make docker-up` | Iniciar PostgreSQL e Redis |
| `make docker-up-full` | Iniciar todos os serviços (incluindo LocalStack) |
| `make docker-status` | Verificar status dos serviços |
| `make docker-logs` | Ver logs dos serviços |
| `make docker-down` | Parar todos os serviços |
| `make docker-clean` | Limpar tudo |

## 🏗️ Arquitetura do Sistema

```
┌─────────────────────────────────────────────────────────────┐
│                     CLIENT REQUESTS                        │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────┴───────────────────────────────────────┐
│                API GATEWAY                                  │
│  ├─ REST API: kivaplus-api                                  │
│  ├─ Routes: /register, /login, /profile, /condominios      │
│  └─ JWT validation handled internally by each Lambda       │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────┴───────────────────────────────────────┐
│                 LAMBDA FUNCTIONS                           │
│  ├─ kivaplus-auth: /register, /login, /refresh             │
│  ├─ kivaplus-users: /profile, /usuarios                    │
│  └─ kivaplus-condominiums: /condominios                    │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────┴───────────────────────────────────────┐
│                   DATA LAYER                               │
│  ├─ PostgreSQL: User data, condominiums, permissions       │
│  ├─ Redis: Session cache, JWT blacklist                    │
│  └─ LocalStack: AWS services simulation (opcional)         │
└─────────────────────────────────────────────────────────────┘
```

## 🔗 Endpoints da API

### Públicos (sem autenticação):
- `POST /register` - Registro de usuário
- `POST /login` - Login de usuário
- `POST /refresh` - Renovar token

### Protegidos (requer JWT):
- `GET /profile` - Obter perfil do usuário
- `PUT /profile` - Atualizar perfil do usuário
- `GET /condominios` - Listar condomínios
- `POST /condominios` - Criar condomínio
- `GET /condominios/{id}` - Obter condomínio específico

## 🧪 Testando a API

### Teste Básico (sem LocalStack):
```bash
# Iniciar apenas PostgreSQL e Redis
make docker-up

# Executar testes unitários
make test

# Executar testes de integração
make test-integration
```

### Teste Completo (com LocalStack):
```bash
# Setup completo incluindo LocalStack
make setup

# Testar via API Gateway
curl -X POST http://localhost:4566/restapis/4gk5dposvy/local/_user_request_/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "João Silva",
    "email": "joao@example.com",
    "password": "MinhaSenh@123",
    "confirm_password": "MinhaSenh@123"
  }'
```

## 🔧 Desenvolvimento

### Estrutura do Projeto:
```
kivaplus-backend/
├── cmd/                   # Entry points das Lambdas
│   ├── auth/             # Lambda de autenticação
│   ├── users/            # Lambda de usuários
│   └── condominiums/     # Lambda de condomínios
├── internal/             # Código interno
│   ├── shared/           # Código compartilhado
│   ├── users/            # Domínio de usuários
│   └── condominiums/     # Domínio de condomínios
├── tests/                # Testes
│   └── integration/      # Testes de integração
├── dist/                 # Artefatos de build
├── docker-compose.yml    # Configuração dos serviços
└── Makefile             # Comandos de desenvolvimento
```

### Modificando o Código:

1. **Altere o código** em `internal/` ou `cmd/`
2. **Reconstrua as Lambdas**: `make build-all`
3. **Teste as mudanças**: `make test-all`

### Adicionando Novos Endpoints:

1. Adicione o handler em `internal/*/handlers/`
2. Implemente o use case em `internal/*/usecases/`
3. Atualize as rotas em `configs/routes.json`
4. Reconstrua: `make build-all`

## 🐳 Serviços Docker

### Serviços Principais:
- **PostgreSQL**: Banco de dados principal
- **Redis**: Cache e sessões
- **LocalStack**: Simulação AWS (opcional)

### Profiles Disponíveis:
```bash
# Build das Lambdas
docker compose --profile build run --rm lambda-builder

# LocalStack (opcional)
docker compose --profile localstack up -d

# Testes
docker compose --profile test run --rm integration-tests

# Setup do LocalStack
docker compose --profile setup run --rm localstack-setup

# Aplicação Go local
docker compose --profile app up -d
```

## 🐛 Solução de Problemas

### Problema: "Connection refused"
```bash
# Verifique se os serviços estão rodando
make docker-status

# Se não estiverem, inicie-os
make docker-up
```

### Problema: "Database connection failed"
```bash
# Verifique o PostgreSQL
docker compose logs postgres

# Reinicie se necessário
docker compose restart postgres
```

### Problema: Testes falhando
```bash
# Execute diagnóstico
make docker-status

# Limpe e reconfigure tudo
make docker-clean
make setup
```

## 📊 Monitoramento

### Logs dos Serviços:
```bash
# Todos os logs
make docker-logs

# Logs específicos
docker compose logs postgres
docker compose logs redis
docker compose logs localstack
```

### Status dos Serviços:
```bash
# Status completo
make docker-status

# Status do Docker
docker compose ps

# Saúde do LocalStack (se rodando)
curl http://localhost:4566/_localstack/health
```

## 🔒 Segurança

- **JWT Tokens**: Expiram em 1 hora
- **Refresh Tokens**: Expiram em 30 dias
- **Senhas**: Hash bcrypt com salt
- **Validação**: Todos os inputs são validados
- **Permissões**: Sistema baseado em roles

## 🚀 Deploy para Produção

Para deploy em AWS real:

1. Configure credenciais AWS
2. Use RDS para PostgreSQL
3. Use ElastiCache para Redis
4. Configure domínio personalizado no API Gateway
5. Deploy via CDK ou Terraform

## 📚 Recursos Adicionais

- [Documentação da API](./docs/)
- [Testes](./tests/)
- [Migrações](./internal/database/migrations/)

## 🤝 Contribuindo

1. Fork o projeto
2. Crie uma branch: `git checkout -b feature/nova-funcionalidade`
3. Commit suas mudanças: `git commit -m 'Adiciona nova funcionalidade'`
4. Push para a branch: `git push origin feature/nova-funcionalidade`
5. Abra um Pull Request

## 📄 Licença

Este projeto está sob a licença Apache 2.0. Veja o arquivo [LICENSE](LICENSE) para detalhes.

## 🆘 Suporte

- **Issues**: [GitHub Issues](https://github.com/kivaplus/kivaplus-backend/issues)
- **Discussões**: [GitHub Discussions](https://github.com/kivaplus/kivaplus-backend/discussions)

---

**Desenvolvido com ❤️ pela equipe KivaPlus**
