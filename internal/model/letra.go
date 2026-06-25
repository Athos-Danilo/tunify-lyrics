package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// StatusLetra define os estados possíveis de processamento
type StatusLetra string

const (
	StatusPendente      StatusLetra = "PENDENTE"
	StatusProcessando   StatusLetra = "PROCESSANDO"
	StatusConcluido     StatusLetra = "CONCLUIDO"
	StatusNaoEncontrada StatusLetra = "NAO_ENCONTRADA"
)

// Letra espelha a coleção de Letras no MongoDB
type Letra struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Artista      string             `bson:"nome_artista" json:"nome_artista"`
	Titulo       string             `bson:"nome_musica" json:"nome_musica"`
	Conteudo     string             `bson:"texto_letra,omitempty" json:"texto_letra,omitempty"`
	Sincronizada bool               `bson:"sincronizada" json:"sincronizada"`
	Fonte        string             `bson:"fonte_letra" json:"fonte_letra"`
	Status       StatusLetra        `bson:"status" json:"status"`
	CriadoEm     time.Time          `bson:"criado_em" json:"criado_em"`
	AtualizadoEm time.Time          `bson:"atualizado_em" json:"atualizado_em"`
}
