package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Letra struct {
	ID                      primitive.ObjectID `bson:"_id,omitempty"`            // ID único da letra no MongoDB
	IDMusicaSpotify         string             `bson:"id_musica_spotify"`        // ID da música no Spotify
	NomeMusica              string             `bson:"nome_musica"`              // Título da música
	NomeArtista             string             `bson:"nome_artista"`             // Nome do artista
	Status                  string             `bson:"status"`                   // Status: PENDENTE, PROCESSANDO, CONCLUIDO, NAO_ENCONTRADA
	TextoLetra              *string            `bson:"texto_letra"`              // Conteúdo da letra (pode ser nulo inicialmente)
	Sincronizada            bool               `bson:"sincronizada"`             // Indica se a letra foi sincronizada
	FonteLetra              *string            `bson:"fonte_letra"`              // Origem/fonte da letra (pode ser nulo)
	TentativasProcessamento int                `bson:"tentativas_processamento"` // Quantidade de tentativas de processamento
	CriadoEm                time.Time          `bson:"criado_em"`                // Data e hora de criação
	AtualizadoEm            time.Time          `bson:"atualizado_em"`            // Data e hora da última atualização
}
