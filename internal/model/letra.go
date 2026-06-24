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
	IDUsuario    primitive.ObjectID `bson:"id_usuario,omitempty" json:"id_usuario,omitempty"`
	Artista      string             `bson:"artista" json:"artista"`
	Titulo       string             `bson:"titulo" json:"titulo"`
	Conteudo     string             `bson:"conteudo,omitempty" json:"conteudo,omitempty"`
	Sincronizada bool               `bson:"sincronizada" json:"sincronizada"`
	Status       StatusLetra        `bson:"status" json:"status"`
	CriadoEm     time.Time          `bson:"criado_em" json:"criado_em"`
	AtualizadoEm time.Time          `bson:"atualizado_em" json:"atualizado_em"`
}
