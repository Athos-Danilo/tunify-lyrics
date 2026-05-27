# 📻 Tunify Letras API — Microserviço de Processamento Assíncrono em Go

[![Go Version](https://img.shields.io/badge/Go-1.20+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Database](https://img.shields.io/badge/Database-MongoDB-47A248?style=flat&logo=mongodb)](https://www.mongodb.com/)
[![Ecossistema](https://img.shields.io/badge/Ecossistema-Tunify-0285FF?style=flat)](https://github.com/athosdanilo/tunify)

Este é o microserviço oficial de captura, processamento e sincronização de letras musicais do ecossistema **Tunify** — uma plataforma de streaming e player musical em constante evolução, projetada como um portfólio vivo de arquitetura de software de alta performance.

Diferente da concepção inicial de busca síncrona (onde o usuário esperava o scraping em tempo real), esta versão implementa uma **Arquitetura Orientada a Filas em Segundo Plano (Worker Pattern)** integrada ao **MongoDB**, otimizada especificamente para operar de forma extremamente econômica em ambientes de hospedagem gratuitos (Free Tier).

---

## 🎯 O que é o Tunify?

O **Tunify** é uma aplicação completa de player musical web e mobile que se conecta a serviços de música para fornecer uma experiência de usuário premium, minimalista e fluida. Ele engloba desde o gerenciamento de filas de reprodução inteligentes até interfaces imersivas baseadas na identidade visual das capas dos álbuns (como o sistema de extração de cores dinâmicas). 

A **Tunify Letras API** nasce para preencher a única lacuna deixada pelas APIs tradicionais de players: o fornecimento de letras de música estruturadas e, sempre que possível, perfeitamente sincronizadas com o tempo da reprodução (estilo Karaokê).

---

## 💡 Arquitetura do Sistema e Design Dinâmico

Para garantir resiliência, baixo consumo de infraestrutura e performance instantânea para o usuário final, o microserviço foi reestruturado sob os seguintes pilares de engenharia:

### 1. Consumo Instantâneo vs. Processamento Assíncrono (Trade-off)
Fazer varreduras em sites externos em tempo real gera latência e instabilidade. Na nova arquitetura, o aplicativo Angular nunca espera o robô fazer o scraping. 
* Se uma música nunca foi tocada na plataforma antes, ela entra na fila com o estado `PENDENTE`. O front-end exibe uma mensagem amigável e honesta informando que o processamento foi agendado.
* A partir da segunda vez que qualquer usuário tocar a mesma música no ecossistema Tunify, a letra já estará disponível para consumo imediato diretamente do banco de dados, carregando em menos de 10 milissegundos.

### 2. O Robô com "Horário de Expediente" (Cron Job com Lock de Segurança)
Para evitar o consumo contínuo e desnecessário de CPU em plataformas como Render ou Vercel (que limitam as horas de execução em planos gratuitos), o robô em Go não roda em loop infinito na nuvem.
* Ele é despertado periodicamente (ex: a cada hora) através de um gatilho agendado (Cron Job).
* **Mecanismo de Lock (Trava):** Para evitar condições de corrida (Race Conditions) — como o robô acordar novamente enquanto o ciclo anterior ainda está processando uma fila longa —, a primeira ação do Go é verificar e criar uma "Trava de Processamento" no banco. Se o robô anterior ainda estiver ativo, a nova instância encerra sua execução pacificamente, protegendo a integridade dos dados e economizando processamento.

### 3. Estratégia de Busca em Cascata (Degradação Suave / Fallback)
O robô foi projetado para ser um caçador resiliente, operando em três níveis de prioridade para cobrir tanto o cenário de músicas globais quanto o acervo nacional/regional brasileiro:
1. **Nível Ouro (LRCLIB):** O robô tenta buscar a letra na API da LRCLIB. Se encontrar, extrai o texto com os metadados de milissegundos (`[00:12.50]Texto`), salvando no banco com a flag `sincronizada: true` para ativar a interface estilo Karaokê no Angular.
2. **Nível Prata (Letras.mus.br / Genius):** Caso a música não possua sincronia na LRCLIB (muito comum em faixas nacionais específicas), o robô realiza o *web scraping* tradicional em portais brasileiros consolidando o texto limpo, salvando com a flag `sincronizada: false` (exibição em modo de rolagem simples).
3. **Nível Bronze (Não Encontrada):** Se após 3 tentativas o robô falhar em todas as fontes (músicas estritamente instrumentais ou inválidas), o status é alterado para `NAO_ENCONTRADA`, impedindo que o sistema gaste recursos tentando processar uma faixa impossível repetidamente.

---

## 🗃️ Estrutura do Banco de Dados (Schema MongoDB)

A coleção no MongoDB chama-se `Letras` e foi totalmente padronizada em português para garantir máxima clareza semântica durante o desenvolvimento. Cada documento segue rigorosamente o modelo abaixo:

```json
{
  "_id": "ObjectId('...')",
  "id_musica_spotify": "3yfqSUWxFvZELEM4PmlwIR",
  "nome_musica": "Yellow",
  "nome_artista": "Coldplay",
  "status": "PENDENTE",
  "texto_letra": null,
  "sincronizada": false,
  "fonte_letra": null,
  "tentativas_processamento": 0,
  "criado_em": "2026-05-27T13:00:00Z",
  "atualizado_em": "2026-05-27T13:00:00Z"
}
```

## Ciclo de Vida dos Estados (`status`):

* `PENDENTE`: Música recém-identificada que aguarda a próxima ativação do robô.
* `PROCESSANDO`: O robô capturou a faixa da fila e está realizando as requisições externas (*lock* a nível de documento).
* `CONCLUIDO`: Letra extraída com sucesso e pronta para o consumo.
* `NAO_ENCONTRADA`: Esgotadas as tentativas de busca nas fontes mapeadas.

## 🛠️ Tecnologias Utilizadas
* **Linguagem Principal:** Go (Golang) — Escolhida pela performance, concorrência nativa e geração de binários minúsculos ideais para microserviços.
* **Web Scraping & Parsing:** `goquery` — Navegação idiomática no DOM HTML inspirada na sintaxe do jQuery.
* **Banco de Dados:** MongoDB (Driver Oficial `go.mongodb.org/mongo-driver`).
* **Comunicação de Rede:** Biblioteca padrão `net/http` do Go, mantendo o serviço enxuto e livre de *frameworks* pesados.
