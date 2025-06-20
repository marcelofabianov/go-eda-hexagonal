# Exemplo projeto GO

O projeto é construído sobre uma arquitetura de backend monolítica modular, seguindo os princípios do Domain-Driven Design (DDD), Arquitetura Hexagonal (Ports & Adapters) e Event-Driven Architecture (EDA). Essa abordagem busca equilibrar a agilidade no desenvolvimento inicial com a flexibilidade e escalabilidade para o futuro.

A stack tecnológica principal é centrada em Go (versão 1.24.x), escolhida pela sua performance, concorrência e robustez. Para o armazenamento de dados, utiliza-se PostgreSQL, um banco de dados relacional confiável, com identificadores UUID v7 para otimização de performance. A persistência de dados inclui estratégias de soft delete e arquivamento lógico para gestão do ciclo de vida dos registros.

A comunicação é gerenciada por uma API RESTful utilizando o roteador go-chi/chi. A Observabilidade é um pilar fundamental, com o OpenTelemetry (OTEL) instrumentando a aplicação para coletar traces distribuídos, que são visualizados no Jaeger. O logging estruturado é feito com log/slog, correlacionado aos traces para facilitar a depuração.

Para a comunicação assíncrona e o backbone da EDA, o projeto emprega NATS JetStream como broker de eventos, configurado em cluster para alta disponibilidade. Ferramentas como go.uber.org/dig para injeção de dependências, pressly/goose para migrações de banco de dados, e golangci-lint com Git Hooks garantem a qualidade e consistência do código.
