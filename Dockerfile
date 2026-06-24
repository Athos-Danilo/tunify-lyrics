# Etapa 1: Build
FROM golang:1.22-alpine AS builder

# Instala certificados SSL e timezone (necessário para a imagem scratch depois e https scraping)
RUN apk add --no-cache ca-certificates tzdata

# Cria usuário não-root para segurança máxima
RUN adduser -D -g '' -H -s /sbin/nologin appuser

WORKDIR /app

# Faz o download das dependências isoladamente (aproveita cache do Docker)
COPY go.mod go.sum ./
RUN go mod download

# Copia o código fonte
COPY . .

# Compila o binário estaticamente otimizado
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /go/bin/tunify-letras ./cmd/server/main.go

# Etapa 2: Imagem Final (Minimalista)
FROM scratch

# Copia os certificados SSL, timezone e o usuário da imagem builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/passwd /etc/passwd

# Copia o binário final
COPY --from=builder /go/bin/tunify-letras /app/tunify-letras

# Troca para o usuário sem privilégios root
USER appuser

# Declara porta padrão (para visualização apenas)
EXPOSE 8080

# Comando de execução
ENTRYPOINT ["/app/tunify-letras"]
