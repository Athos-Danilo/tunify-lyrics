# Tunify Lyrics API — Microserviço em Go

[![Go Version](https://img.shields.io/badge/Go-1.20+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Microservice](https://img.shields.io/badge/Ecossistema-Tunify-0285FF?style=flat)]()

Este é um microserviço de alta performance desenvolvido em **Go (Golang)**, projetado especificamente para atuar como o motor de busca e extração de letras musicais da aplicação **Tunify**.

---

## 🎯 O que este serviço faz?

A **Tunify Lyrics API** resolve um desafio de integração com plataformas de música: a obtenção automatizada e limpa de letras. 

Como a API oficial do Genius não fornece o texto estruturado das letras devido a direitos autorais (apenas os metadados e o link da página), este microserviço assume a responsabilidade de:
1. Receber as requisições do front-end do **Tunify** contendo o nome da música e do artista.
2. Identificar e extrair cirurgicamente a estrutura HTML correta diretamente do ecossistema Genius através de *Web Scraping* em tempo real.
3. Formatar, limpar e processar o texto, devolvendo uma resposta JSON leve e estruturada, pronta para ser consumida pela interface do player.

---

## 💡 Motivação e Arquitetura

O **Tunify** é um ecossistema musical em constante evolução, funcionando como um *portfólio vivo* para consolidar e demonstrar conceitos avançados de engenharia de software. A decisão de isolar a busca de letras em um microserviço próprio e em uma nova tecnologia foi guiada por três pilares fundamentais:

* **Isolamento de Falhas e Resiliência:** Rotinas de *Web Scraping* são inerentemente frágeis, pois dependem da estrutura visual de terceiros. Ao isolar essa lógica em um microserviço independente, garantimos que qualquer mudança repentina no layout do Genius afete apenas a funcionalidade de letras, mantendo o coração da aplicação principal totalmente estável e online.
* **Poliglotismo de Infraestrutura:** Expandindo os horizontes técnicos para além do ecossistema Python e TypeScript já utilizados no projeto principal, a introdução do **Go** adiciona maturidade e versatilidade ao portfólio, demonstrando a habilidade de construir aplicações eficientes em diferentes paradigmas de programação.
* **Performance Extrema e Baixo Custo:** O Go compila para um único binário nativo, entregando tempos de resposta na casa dos milissegundos com consumo de memória RAM insignificante. Isso o torna a escolha perfeita para arquiteturas *cloud-native* e hospedagens eficientes em plataformas na nuvem.

---

## 🛠️ Tecnologias Utilizadas

* **Linguagem Principal:** Go (Golang) — Eficiência, tipagem forte e concorrência nativa.
* **Web Scraping:** `goquery` — Biblioteca robusta baseada na sintaxe do jQuery para navegação e extração de nós do DOM HTML.
* **Roteamento:** Módulo `net/http` nativo do Go para construção de APIs enxutas.

---

## 🚀 Como Executar o Projeto Localmente

### Pré-requisitos
* Ter o [Go instalado](https://go.dev/doc/install) na sua máquina.

### Passos para inicialização
1. Clone este repositório:
   ```bash
   git clone [https://github.com/SEU_USUARIO_AQUI/tunify-lyrics-api.git](https://github.com/SEU_USUARIO_AQUI/tunify-lyrics-api.git)
   cd tunify-lyrics-api