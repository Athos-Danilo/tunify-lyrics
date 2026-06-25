package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CotaDiaria rastreia a quantidade de letras buscadas no dia atual para evitar bloqueios em Cloud / Free Tier
type CotaDiaria struct {
	ID                 primitive.ObjectID            `bson:"_id,omitempty" json:"id,omitempty"`
	Data               string                        `bson:"data" json:"data"` // Ex: "2026-06-24"
	ContagemGlobal     int                           `bson:"contagem_global" json:"contagem_global"`
}
