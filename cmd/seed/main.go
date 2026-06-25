package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/athosdanilo/tunify-letras/internal/config"
	"github.com/athosdanilo/tunify-letras/internal/db"
	"github.com/athosdanilo/tunify-letras/internal/logger"
	"github.com/athosdanilo/tunify-letras/internal/model"
)

func main() {
	logger.Init()
	if err := config.Load(); err != nil {
		log.Fatal(err)
	}

	database, err := db.GetDatabase()
	if err != nil {
		log.Fatal(err)
	}

	col := database.Collection("letras")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Inserir 3 musicas pendentes de teste

	musicas := []interface{}{
		model.Letra{
			Artista:   "Coldplay",
			Titulo:    "Yellow",
			Status:    model.StatusPendente,
		},
		model.Letra{
			Artista:   "Henrique e Juliano",
			Titulo:    "Cuida Bem Dela",
			Status:    model.StatusPendente,
		},
		model.Letra{
			Artista:   "UmaBandaFalsa123",
			Titulo:    "MusicaQueNaoExiste",
			Status:    model.StatusPendente,
		},
	}

	res, err := col.InsertMany(ctx, musicas)
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("Seed realizado com sucesso! Inseridos %d documentos no banco de dados 'tunificar', coleção 'letras'!\n", len(res.InsertedIDs))
}
