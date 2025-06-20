#!/bin/bash

# =================================================================
# Docker & Application Aliases
# =================================================================
alias g="docker exec -it redtogreen-api"
alias gl="docker compose logs -f redtogreen-api"
alias gd="docker compose up -d"
alias gb="g bash"
alias gds="docker compose stats"

# =================================================================
# Go Tooling Aliases
# =================================================================
alias gg="g go"
alias gr="gg run"
alias gtest="gg test"

# =================================================================
# Goose Migration Aliases
# =================================================================
alias gs="g goose"
alias gup="gs up"
alias gdown="gs down"
alias greset="gs reset"
alias gcreate="gs create"

alias gauditdb_up='docker-compose exec \
    -e GOOSE_MIGRATION_DIR="/app/db/migrations_audit" \
    -e GOOSE_DBSTRING="postgres://${APP_AUDIT_DB_USER:-audituser}:${APP_AUDIT_DB_PASSWORD:-auditpass}@${APP_AUDIT_DB_HOST:-redtogreen-audit-db}:${APP_AUDIT_DB_PORT:-5433}/${APP_AUDIT_DB_NAME:-redtogreen-audit-db}?sslmode=${APP_AUDIT_DB_SSL_MODE:-disable}" \
    redtogreen-api goose up'

# =================================================================
# NATS Tooling Aliases
# Todos os comandos agora usam 'docker run' com 'nats-box' para garantir
# que a CLI do NATS esteja sempre disponível, sem poluir a imagem da API.
# =================================================================

# Inicia um container nats-box para interação manual
alias gnats-box="docker run -it --rm --network redtogreen_redtogreen-network natsio/nats-box"

# Função para buscar uma mensagem específica pelo ID da sequência.
# Uso: gnats-get <stream_name> <sequence_id>
# Exemplo: gnats-get identity-stream 1
gnats-get() {
  if [ -z "$1" ] || [ -z "$2" ]; then
    echo "Usage: gnats-get <stream_name> <sequence_id>"
    return 1
  fi
  docker run --rm --network redtogreen_redtogreen-network natsio/nats-box \
    nats stream get "$1" "$2" -s nats://nats-0:4222
}

# Função para buscar as N últimas mensagens de um stream, em ordem da mais nova para a mais antiga.
# Uso: gnats-last <stream_name> [number_of_messages]
# Exemplo 1 (apenas a última): gnats-last identity-stream
# Exemplo 2 (as últimas 5):   gnats-last identity-stream 5
gnats-last() {
  if [ -z "$1" ]; then
    echo "Usage: gnats-last <stream_name> [number_of_messages]"
    return 1
  fi
  local stream_name=$1
  local count=${2:-1}
  local info_output=$(docker run --rm --network redtogreen_redtogreen-network natsio/nats-box nats stream info "$stream_name" --json -s nats://nats-0:4222 2>/dev/null)
  if [ -z "$info_output" ]; then
    echo "Error: Could not get info for stream '$stream_name'."
    return 1
  fi
  local last_seq=$(echo "$info_output" | grep -o '"last_seq": *[0-9]*' | awk '{print $2}')
  if [ -z "$last_seq" ] || [ "$last_seq" -eq 0 ]; then
    echo "Stream '$stream_name' is empty."
    return 0
  fi
  local start_seq=$((last_seq - count + 1))
  if [ "$start_seq" -lt 1 ]; then
    start_seq=1
  fi
  echo "--> Fetching last $count message(s) from stream '$stream_name' (sequences $last_seq down to $start_seq)..."
  local script=""
  for i in $(seq "$last_seq" -1 "$start_seq"); do
    script+="nats stream get '$stream_name' '$i' -s nats://nats-0:4222;"
    if [ "$i" -ne "$start_seq" ]; then
      script+="echo '--------------------------------------------------';"
    fi
  done
  docker run -it --rm --network redtogreen_redtogreen-network natsio/nats-box /bin/sh -c "$script"
}

# Função para apagar (fazer purge) de todas as mensagens de um stream.
# Uso: gnats-purge <stream_name>
# Exemplo: gnats-purge identity-stream
gnats-purge() {
    if [ -z "$1" ]; then
        echo "Usage: gnats-purge <stream_name>"
        return 1
    fi
    echo "--> Purging all messages from stream '$1'..."
    docker run --rm --network redtogreen_redtogreen-network natsio/nats-box \
        nats stream purge "$1" -f -s nats://nats-0:4222
}

# Função para inspecionar os detalhes de um consumidor específico.
# Uso: gnats-consumer <stream_name> <consumer_name>
# Exemplo: gnats-consumer identity-stream user.created-processor
gnats-consumer() {
    if [ -z "$1" ] || [ -z "$2" ]; then
        echo "Usage: gnats-consumer <stream_name> <consumer_name>"
        return 1
    fi
    echo "--> Getting info for consumer '$2' on stream '$1'..."
    docker run -it --rm --network redtogreen_redtogreen-network natsio/nats-box \
        nats consumer info "$1" "$2" -s nats://nats-0:4222
}
