## Exemplo de Estrutura de Evento JSON

Este é um exemplo da estrutura de um evento `user.created`, refletindo o design que estamos trabalhando:

```json
{
  "header": {
    "eventId": "a7b8c9d0-e1f2-3g4h-5i6j-7k8l9m0n1o2p",
    "eventType": "user.created",
    "timestamp": "2025-06-19T10:00:00Z",
    "source": "IdentityService",
    "jsonSchemaVersion": "v1.0.0"
  },
  "context": {
    "correlationId": "xyz123abc456-transacao-inicial",
    "userId": "uuid-usuario-que-executou-a-acao"
  },
  "metadata": {
    "traceId": "trace-id-da-execucao-otel",
    "previousEventId": null,
    "causationId": null
  },
  "payload": {
    "userId": "uuid-usuario-criado",
    "name": "Maria Silva",
    "email": "maria.silva@example.com",
    "phone": "5511987654321"
  }
}
```

## Legenda dos Campos do JSON

### `header`

Contém informações essenciais sobre o evento em si.

* **`eventId` (UUID/GUID):** Um identificador único e imutável para *este evento específico*. É fundamental para idempotência, rastreamento e auditoria.

    * *Exemplo:* `a7b8c9d0-e1f2-3g4h-5i6j-7k8l9m0n1o2p`

* **`eventType` (String):** O tipo de evento que ocorreu, geralmente no formato `domínio.ação`. Define a natureza da mudança de estado.

    * *Exemplo:* `user.created`, `order.placed`, `payment.processed`

* **`timestamp` (ISO 8601 String):** O momento exato (data e hora em UTC) em que o evento foi gerado pelo serviço de origem. Essencial para ordenação cronológica e análises temporais.

    * *Exemplo:* `2025-06-19T10:00:00Z`

* **`source` (String):** O nome do serviço, microsserviço ou componente que foi o *produtor* (publicador) deste evento. Ajuda a identificar a origem e responsabilidade.

    * *Exemplo:* `IdentityService`, `OrderService`, `PaymentService`

* **`jsonSchemaVersion` (String):** A versão do contrato (schema) do `payload` (e, indiretamente, da estrutura do evento). Permite que os consumidores saibam como interpretar o conteúdo e lidar com evoluções futuras sem quebrar.

    * *Exemplo:* `v1.0.0`, `v1.1.0`

### `context`

Contém informações de contexto de negócio ou transacionais que são relevantes para a operação mais ampla à qual o evento pertence.

* **`correlationId` (UUID/GUID):** Um identificador que percorre uma cadeia completa de operações distribuídas. É o mesmo ID para a solicitação inicial e todos os eventos subsequentes disparados por ela. Permite rastrear uma transação completa através de múltiplos serviços e eventos.

    * *Exemplo:* `xyz123abc456-transacao-inicial`

* **`userId` (UUID/GUID, Opcional):** O identificador do usuário que iniciou a ação que resultou neste evento, se aplicável. Útil para auditoria e contexto de segurança.

    * *Exemplo:* `user-456-quem-executou` (pode ser `null` se a ação foi iniciada por um sistema ou processo automático)

### `metadata`

Contém metadados técnicos ou de encadeamento de eventos, importantes para observabilidade e depuração.

* **`traceId` (UUID/GUID):** O ID da "trace" (rastreamento) para sistemas de observabilidade como OpenTelemetry ou Jaeger. Permite correlacionar este evento a uma trace distribuída que abrange múltiplas operações e serviços.

    * *Exemplo:* `trace-id-da-execucao-otel`

* **`previousEventId` (UUID/GUID, Opcional):** O `eventId` do evento que foi publicado imediatamente antes deste na sequência lógica de um fluxo. Ajuda a reconstruir a ordem dos eventos.

    * *Exemplo:* `null` (se for o primeiro evento em uma sequência) ou um UUID de um evento anterior.

* **`causationId` (UUID/GUID, Opcional):** O `eventId` do evento que *causou diretamente* este evento. Por exemplo, um `PaymentProcessed` pode ser o `causationId` para um `OrderShipped`. Ajuda a entender as relações de causalidade entre eventos.

    * *Exemplo:* `null` (se não for causado por outro evento específico) ou um UUID do evento causador.

### `payload`

O corpo principal do evento, contendo os dados específicos do domínio que representam a mudança de estado que o evento está comunicando. O conteúdo do `payload` deve ser imutável após a publicação.

* **`userId` (UUID/GUID):** O ID do usuário recém-criado.
* **`name` (String):** O nome completo do usuário.
* **`email` (String):** O endereço de e-mail do usuário.
* **`phone` (String):** O número de telefone do usuário.
