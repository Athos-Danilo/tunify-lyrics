# Nome do executável
APP_NAME=tunify-letras

# Variáveis do Go
GO=go
GOBUILD=$(GO) build
GOTEST=$(GO) test
GOVET=$(GO) vet
GOCLEAN=$(GO) clean

.PHONY: all build run test lint docker clean help

# Comando padrão
all: build

## build: Compila o projeto
build:
	@echo "Compilando $(APP_NAME)..."
	$(GOBUILD) -o $(APP_NAME) ./cmd/server/main.go
	@echo "Compilação concluída!"

## run: Roda a aplicação localmente
run:
	@echo "Rodando $(APP_NAME)..."
	$(GO) run ./cmd/server/main.go

## test: Roda todos os testes unitários
test:
	@echo "Rodando testes..."
	$(GOTEST) -v ./...

## lint: Roda o go vet (ou golangci-lint se instalado)
lint:
	@echo "Rodando linter..."
	$(GOVET) ./...

## docker: Faz o build da imagem Docker Multi-stage
docker:
	@echo "Construindo imagem Docker..."
	docker build -t $(APP_NAME):latest .

## clean: Remove binários compilados
clean:
	@echo "Limpando projeto..."
	$(GOCLEAN)
	rm -f $(APP_NAME)

## help: Mostra os comandos disponíveis
help:
	@echo "Comandos disponíveis:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':'
