package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Letra representa o documento exato que vai morar no nosso MongoDB
type Letra struct {
	ID                      primitive.ObjectID `bson:"_id,omitempty"`
	IDMusicaSpotify         string             `bson:"id_musica_spotify"`
	NomeMusica              string             `bson:"nome_musica"`
	NomeArtista             string             `bson:"nome_artista"`
	Status                  string             `bson:"status"`      // PENDENTE, PROCESSANDO, CONCLUIDO, NAO_ENCONTRADA
	TextoLetra              *string            `bson:"texto_letra"` // Ponteiro (*) porque pode ser nulo no começo
	Sincronizada            bool               `bson:"sincronizada"`
	FonteLetra              *string            `bson:"fonte_letra"`
	TentativasProcessamento int                `bson:"tentativas_processamento"`
	CriadoEm                time.Time          `bson:"criado_em"`
	AtualizadoEm            time.Time          `bson:"atualizado_em"`
}
