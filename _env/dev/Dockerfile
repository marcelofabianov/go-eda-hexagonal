FROM golang:1.24-alpine

ARG HOST_UID=1000
ARG HOST_GID=1000

RUN addgroup -g ${HOST_GID} -S appgroup && \
  adduser -u ${HOST_UID} -G appgroup -S -D -h /home/appuser appuser

RUN mkdir -p /app && chown -R appuser:appgroup /app
WORKDIR /app

RUN apk add --no-cache bash curl git unzip && \
  go install github.com/pressly/goose/v3/cmd/goose@latest && \
  go install github.com/air-verse/air@latest

ENV PATH=$PATH:/go/bin

RUN chown -R appuser:appgroup /go

USER appuser

COPY --chown=appuser:appgroup go.mod go.sum ./
RUN chmod 444 go.mod go.sum && \
  go mod download

EXPOSE 8000 8080
CMD ["air", "-c", ".air.toml"]
