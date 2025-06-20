# Exemplo projeto GO

O projeto é construído sobre uma arquitetura de backend monolítica modular, seguindo os princípios do Domain-Driven Design (DDD), Arquitetura Hexagonal (Ports & Adapters) e Event-Driven Architecture (EDA). Essa abordagem busca equilibrar a agilidade no desenvolvimento inicial com a flexibilidade e escalabilidade para o futuro.

A stack tecnológica principal é centrada em Go (versão 1.24.x), escolhida pela sua performance, concorrência e robustez. Para o armazenamento de dados, utiliza-se PostgreSQL, um banco de dados relacional confiável, com identificadores UUID v7 para otimização de performance. A persistência de dados inclui estratégias de soft delete e arquivamento lógico para gestão do ciclo de vida dos registros.

A comunicação é gerenciada por uma API RESTful utilizando o roteador go-chi/chi. A Observabilidade é um pilar fundamental, com o OpenTelemetry (OTEL) instrumentando a aplicação para coletar traces distribuídos, que são visualizados no Jaeger. O logging estruturado é feito com log/slog, correlacionado aos traces para facilitar a depuração.

Para a comunicação assíncrona e o backbone da EDA, o projeto emprega NATS JetStream como broker de eventos, configurado em cluster para alta disponibilidade. Ferramentas como go.uber.org/dig para injeção de dependências, pressly/goose para migrações de banco de dados, e golangci-lint com Git Hooks garantem a qualidade e consistência do código.

### 💻 Stack Tecnológica (Backend)

* **Linguagem:** Go (versão 1.24.x)
* **Framework Web:** `go-chi/chi`
* **Banco de Dados:** PostgreSQL (principal e auditoria)
* **Broker de Mensagens:** NATS (com JetStream para persistência de eventos)
* **Containerização:** Docker, Docker Compose
* **Observabilidade:** OpenTelemetry (OTEL) com Jaeger para Tracing Distribuído
* **Testes:** `stretchr/testify`
* **Injeção de Dependências:** `go.uber.org/dig`
* **Migrations DB:** `pressly/goose`
* **Configuração:** `spf13/viper`

Para uma descrição mais aprofundada da stack e seus princípios, consulte `_doc/STACK.md`.

---

## 🚀 Como Rodar o Projeto

Para colocar o projeto RedToGreen em funcionamento no seu ambiente de desenvolvimento, siga estes passos:

1.  **Pré-requisitos:**
    * Docker e Docker Compose (compatíveis com Linux)
    * `make` (GNU Make)

2.  **Configure o ambiente de desenvolvimento:**
    * Este comando prepara os arquivos Docker (`Dockerfile`, `docker-compose.yml`, `.env` e `.project_aliases.sh`) na raiz do seu projeto, preenchendo variáveis de ambiente como `HOST_UID`/`HOST_GID`.
    * Execute:
        ```bash
        make setup-dev
        ```

3.  **Instale os Git Hooks:** (Config em andamento)
    * Este passo copia o script `pre-commit.sh` para o diretório de hooks do Git e o torna executável, garantindo verificações de qualidade antes dos commits.
    * Execute:
        ```bash
        make install-git-hooks
        ```

4.  **Inicie os serviços Docker:**
    * Este comando constrói a imagem da API e levanta todos os containers necessários (API, bancos de dados, NATS, Jaeger).
    * Execute:
        ```bash
        docker compose up -d --build
        # Alternativa via alias: gd
        ```
    * Aguarde alguns segundos para todos os serviços estarem completamente operacionais (o hook `pre-commit` já fará uma verificação de prontidão antes de cada commit Go).

5.  **Aplique as migrações do banco de dados:**
    * Isso cria as tabelas necessárias no banco de dados principal e de auditoria.
    * Para o banco de dados principal:
        ```bash
        docker compose exec redtogreen-api goose up
        # Alternativa via alias: gup
        ```
    * Para o banco de dados de auditoria:
        ```bash
        docker compose exec \
            -e GOOSE_MIGRATION_DIR="/app/db/migrations_audit" \
            -e GOOSE_DBSTRING="postgres://audituser:auditpass@redtogreen-audit-db:5433/redtogreen-audit-db?sslmode=disable" \
            redtogreen-api goose up
        # Alternativa via alias: gauditdb_up
        ```

6.  **Verifique se a API está funcionando:**
    * A API estará acessível em `http://localhost:8080`.
    * Você pode verificar o endpoint de saúde:
        ```bash
        curl http://localhost:8080/ping
        ```
        * Você deve receber uma resposta de sucesso (geralmente sem corpo ou com um status 200).

7.  **Teste a criação de um usuário (exemplo instrumentado com OpenTelemetry):**
    * Este é um exemplo completo que testa a funcionalidade, incluindo a persistência no banco e o registro de auditoria via NATS.
    * Execute:
        ```bash
        curl -X POST http://localhost:8080/api/v1/identity/users \
        -H "Content-Type: application/json" \
        -d '{
          "name": "Marcelo Fabiano",
          "email": "marcelo.fabiano@example.com",
          "password": "StrongPassword123!",
          "password_confirmation": "StrongPassword123!",
          "phone": "5511987654321"
        }'
        ```
    * Você deverá receber uma resposta `201 Created` com os dados do novo usuário.
    * Para visualizar os traces desta requisição, acesse o [Jaeger UI](http://localhost:16686) e selecione o serviço `redtogreen-api`.

---

## 🤝 Contribuindo

Seja bem-vindo para contribuir com o RedToGreen! Por favor, consulte os documentos de `ADRs` e `Glossário` para entender as decisões arquiteturais e o vocabulário do projeto.

---

## 📄 Licença

Este projeto está licenciado sob a licença [MIT](LICENSE).

---
