# 📋 Tarefas de Desenvolvimento: Tunify Letras Microservice (Worker & API)

Este documento detalha o plano de ação passo a passo para a construção do microserviço **Tunify Letras** em Go. O objetivo é entregar um sistema **absurdamente rápido, inteligente e que consuma pouquíssimo hardware**, utilizando o padrão ouro do mercado e as melhores práticas de Engenharia de Software.

## 🏆 Padrão Ouro e Melhores Práticas Adotadas
Para garantir a máxima performance com o menor consumo de infraestrutura possível, este projeto seguirá os seguintes princípios:
- **Linguagem e Runtime:** Go (Golang) será utilizado por sua altíssima eficiência de CPU/Memória, suporte nativo a concorrência (Goroutines) e geração de binários minúsculos e estáticos.
- **Arquitetura Limpa (Clean Architecture):** Separação estrita de responsabilidades (Camadas de Domínio, Caso de Uso/Serviço, Repositório e Infraestrutura).
- **Design Patterns Essenciais:**
  - *Singleton/Pool:* Para reaproveitamento da conexão com o MongoDB.
  - *Strategy:* Para encapsular os diferentes motores de busca de letras (LRCLIB, Letras.mus.br, Genius), permitindo fácil expansão no futuro.
- **Resiliência e Tolerância a Falhas:**
  - Implementação de retries com *Exponential Backoff* nas chamadas a APIs externas.
  - Prevenção de travamentos (*Deadlocks*) com uso correto de Contextos (`context.Context`) e Timeouts em todas as requisições HTTP e do Banco de Dados.
- **Alta Performance:** Uso de *Worker Pools* para limitar e otimizar a quantidade máxima de goroutines executadas simultualmente durante os picos de varredura de fila.
- **Empacotamento Cirúrgico:** Dockerfile *Multi-stage build* terminando em uma imagem `scratch` ou `alpine` (garantindo contêineres extremamente leves, geralmente com menos de 20MB).
- **Comentários no Código:** Escrever comentários no código para facilitar o entendimento e a manutenção, garantindo que qualquer pessoa que ler o código entenda a lógica.

---

## 🎯 Épico 1: Setup e Infraestrutura Base
- [x] **Configuração do Projeto Go:**
  - Inicializar o módulo (`go mod init github.com/athosdanilo/tunify-letras`).
  - Estruturar os diretórios (ex: `cmd/`, `internal/`, `pkg/`).
- [x] **Gerenciamento de Configurações:**
  - Implementar leitura e validação de variáveis de ambiente (`.env`) via biblioteca nativa ou `godotenv`.
  - Definir variáveis essenciais (`MONGO_URI`, `DATABASE_NAME`, `CRON_INTERVAL`, `PORT`).
- [x] **Conexão com o Banco de Dados (MongoDB):**
  - Configurar conexão segura via driver `go.mongodb.org/mongo-driver`.
  - Implementar Padrão Singleton para gerir o *Connection Pool* e evitar estouro de conexões.
  - Configurar índices no MongoDB (ex: índice composto em `status` e `atualizado_em` para buscas ultra-rápidas na fila).
- [x] **Logger Estruturado:**
  - Configurar logs no padrão JSON para fácil rastreabilidade no terminal e integração com plataformas de observabilidade, usando a biblioteca nativa `slog` (Go 1.21+) ou `zap`.

## 🎯 Épico 2: Repositório e Gestão de Estado
- [x] **Modelagem de Dados:**
  - Criar o *Struct* Go (`model.Letra`) espelhando com perfeição a coleção de Letras do MongoDB usando as tags `bson`.
  - **Atenção:** Atualizar o *Struct* para garantir o mapeamento do campo `id_usuario` no futuro, para habilitar as restrições de limites mensais.
- [x] **Operações Atômicas de Banco:**
  - Criar método `BuscarMusicaPendente()`.
  - **Mecanismo de Lock Profissional:** Usar o comando `FindOneAndUpdate` do MongoDB para buscar um item com status `PENDENTE` e alterá-lo instantaneamente para `PROCESSANDO` em uma única operação atômica, eliminando totalmente o risco de *Race Conditions* (goroutines ou contêineres diferentes processando a mesma música ao mesmo tempo).
  - Criar método `AtualizarStatusMusica()` para salvar os resultados como `CONCLUIDO` (com ou sem sincronia) ou `NAO_ENCONTRADA`.

## 🎯 Épico 3: Motores de Busca e Scraping Inteligente
- [x] **Implementar Padrão Strategy:**
  - Criar a interface `LyricsProvider` com o contrato `Fetch(artista string, titulo string) (*Result, error)`.
- [x] **Provedor Nível Ouro (API LRCLIB):**
  - Integrar com os endpoints públicos da LRCLIB.
  - Lógica para processar JSON, validar existência de tempo (tags de sincronia ex: `[00:15.30]`) e extrair.
  - Definir a flag `sincronizada: true` no retorno.
- [x] **Provedor Nível Prata (Letras.mus.br Scraping):**
  - Utilizar a biblioteca `goquery`.
  - Desenvolver uma função inteligente de normalização de strings (Regex) para converter nomes ("Coldplay", "Yellow") no formato correto das URLs do site.
  - Realizar o Scraping da Div contendo a letra e remover anúncios ou quebras de linhas desnecessárias.
  - Definir a flag `sincronizada: false` no retorno.
- [x] **Orquestrador em Cascata (Fallback Manager):**
  - Lógica de fluxo: Tenta `Ouro` -> Se falhar ou estiver indisponível -> Tenta `Prata` -> Se falhar -> Exaure as tentativas e encerra como `NAO_ENCONTRADA`.

## 🎯 Épico 4: Motor de Processamento Assíncrono (Worker)
- [ ] **Agendador Cron Embutido (Trabalhador Calmo):**
  - Integrar pacote (como `robfig/cron/v3`) ou usar um Ticker nativo (`time.Ticker`).
  - Configurar para despertar de forma leve a cada 15 a 30 minutos, poupando a nuvem (Free Tier).
- [ ] **Gerenciamento de Fila com Cota e Fila Justa (*Fair Queuing*):**
  - Controlar as cotas usando variáveis de ambiente (ex: 100 globais diárias, limite de 20 por usuário diárias).
  - Desenvolver uma `aggregation pipeline` ou lógica no MongoDB para puxar as tarefas em padrão *Round Robin* por `id_usuario` (2 letras de um, 2 do outro) não deixando um "super usuário" dominar o ciclo do cron.
- [ ] **Proteção de IP e Profilaxia (Jitter/Sleep):**
  - Inserir um `time.Sleep` de cerca de 5 segundos entre cada requisição processada dentro de um lote. Isso previne tomar blocos HTTP `429` ou de WAF (Cloudflare) provenientes do site de letras por "acesso rápido demais".
  - Se tomar `429 Too Many Requests`, acionar o *Exponential Backoff* ou suspender a fila totalmente por horas (Retiro Espiritual).

## 🎯 Épico 5: Interface de Controle (API REST em Go)
- [ ] **Servidor HTTP Extremamente Leve:**
  - Subir servidor web nativo usando `net/http` ou o novo multiplexador do Go 1.22+.
- [ ] **Endpoint `/health`:**
  - Rota `GET` que responde `200 OK` instantâneo. Essencial para contêineres na nuvem (ex: Render, AWS ECS) manterem a aplicação acordada e validarem se não houve Crash.
- [ ] **Endpoint `/trigger` (Acionador Imediato):**
  - Rota `POST` para forçar o Worker a iniciar o processamento da fila imediatamente, útil se o backend em Python (FastAPI) quiser avisar o Go: "Ei, acabei de inserir uma música na fila, acorde agora!".

## 🎯 Épico 6: Garantia de Qualidade e Testes Automáticos
- [ ] **Testes Unitários da Regra de Negócio:**
  - Usar pacotes nativos `testing` do Go.
  - Testar o "Orquestrador em Cascata" (Mockando falhas falsas no nível ouro para garantir que ele cai no nível prata adequadamente).
  - Testar o sanitizador de strings do Scraper do Letras.mus.br sem fazer chamadas HTTP de verdade.
- [ ] **Proteção de Código Analítica (Linter):**
  - Configurar `golangci-lint` para garantir que o padrão ouro da linguagem seja seguido em todos os arquivos de código.

## 🎯 Épico 7: Distribuição e Deploy Minimalista
- [ ] **Dockerização Multi-stage:**
  - Etapa 1: Imagem `golang:alpine`, baixar dependências `go mod`, rodar testes e compilar binário com tags de otimização (`-s -w` e `CGO_ENABLED=0`).
  - Etapa 2: Imagem final baseada puramente em `scratch` ou pacote mínimo `alpine`.
  - Configurar um `USER` sem privilégios dentro do Docker para segurança máxima.
- [ ] **Documentação Extra:**
  - Deixar no repositório comandos úteis de Makefile para rodar o app rapidamente.
