package tasks

import (
	"context"
	"fmt"
	"time"

	"tunify-lyrics-api/db"
	"tunify-lyrics-api/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// IniciarRobo liga o motor do nosso Worker
func IniciarRobo() {
	fmt.Println("🤖 Robô operário ativado! Trabalhando em lotes...")

	for {
		fmt.Println("\n⏳ [ROBÔ] Iniciando novo ciclo de processamento da fila...")
		processarLote()

		fmt.Println("💤 [ROBÔ] Lote finalizado ou fila vazia. Indo dormir...")

		// 🚨 PARA TESTES AGORA: Vamos deixar ele dormir só 1 minuto para você ver funcionando.
		// Quando for mandar para produção, troque para: time.Sleep(1 * time.Hour)
		time.Sleep(1 * time.Minute)
	}
}

func processarLote() {
	limiteLote := 15

	// O robô vai tentar puxar até 15 músicas da fila
	for i := 1; i <= limiteLote; i++ {

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		// 1. A BUSCA COM LOCK
		filtro := bson.M{"status": "PENDENTE"}
		atualizacao := bson.M{
			"$set": bson.M{
				"status":        "PROCESSANDO",
				"atualizado_em": time.Now(),
			},
		}

		opcoes := options.FindOneAndUpdate().SetReturnDocument(options.After)

		var letra models.Letra
		err := db.LetrasCollection.FindOneAndUpdate(ctx, filtro, atualizacao, opcoes).Decode(&letra)

		// Libera o cronômetro do banco IMEDIATAMENTE após a busca, para economizar memória
		cancel()

		if err == mongo.ErrNoDocuments {
			// A fila está vazia! O robô percebe e usa o "break" para quebrar o loop das 15 tentativas.
			fmt.Println("📭 [ROBÔ] Nenhuma música PENDENTE encontrada. Fila limpa!")
			break
		} else if err != nil {
			fmt.Println("🚨 [ROBÔ] Erro bizarro ao buscar música na fila:", err)
			break
		}

		fmt.Printf("🔍 [ROBÔ] [%d/%d] Processando: '%s' do '%s'...\n", i, limiteLote, letra.NomeMusica, letra.NomeArtista)

		// =========================================================
		// 🚨 LÓGICA DE SCRAPING ENTRARÁ AQUI NO PRÓXIMO PASSO!
		// =========================================================

		// Simulação: O robô tá navegando na internet...
		time.Sleep(3 * time.Second)

		// 2. SALVA O TRABALHO CONCLUÍDO
		ctxUpdate, cancelUpdate := context.WithTimeout(context.Background(), 5*time.Second)

		textoFalso := "[Verse 1]\nLetra processada no sistema de Lotes com sucesso!\n"
		conclusao := bson.M{
			"$set": bson.M{
				"status":        "CONCLUIDO",
				"texto_letra":   &textoFalso,
				"sincronizada":  false,
				"atualizado_em": time.Now(),
			},
		}

		_, err = db.LetrasCollection.UpdateByID(ctxUpdate, letra.ID, conclusao)
		cancelUpdate() // Libera a memória da atualização

		if err != nil {
			fmt.Println("🚨 [ROBÔ] Erro ao salvar conclusão:", err)
			continue // O "continue" diz: "Deu erro nessa música, mas pule direto para a próxima do lote!"
		}

		fmt.Printf("✅ [ROBÔ] Letra de '%s' salva no banco!\n", letra.NomeMusica)

		// 3. A MÁGICA ANTI-BLOQUEIO (Limitação de Taxa)
		// Se essa ainda não for a última música da fila, ele dá um respiro para enganar o firewall do Genius
		if i < limiteLote {
			fmt.Println("⏱️ [ROBÔ] Pausa estratégica de 8 segundos para evitar bloqueio de IP...")
			time.Sleep(8 * time.Second)
		}
	}
}
