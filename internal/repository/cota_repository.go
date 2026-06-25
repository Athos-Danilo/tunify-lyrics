package repository

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/athosdanilo/tunify-letras/internal/model"
)

// CotaRepository provê as operações de banco de dados para a coleção de cota diária
type CotaRepository struct {
	collection *mongo.Collection
}

// NewCotaRepository cria uma nova instância de CotaRepository
func NewCotaRepository(db *mongo.Database) *CotaRepository {
	return &CotaRepository{
		collection: db.Collection("cotas_diarias"),
	}
}

// ObterCotaDoDia retorna o documento de cota da data especificada.
// Se não existir, ele cria o documento zerado.
func (r *CotaRepository) ObterCotaDoDia(ctx context.Context, data string) (*model.CotaDiaria, error) {
	filter := bson.M{"data": data}
	
	update := bson.M{
		"$setOnInsert": bson.M{
			"data": data,
			"contagem_global": 0,
		},
	}
	
	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)
		
	var cota model.CotaDiaria
	err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&cota)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	}
	
	return &cota, nil
}

// IncrementarCota incrementa a cota global de forma atômica.
func (r *CotaRepository) IncrementarCota(ctx context.Context, data string) error {
	filter := bson.M{"data": data}

	update := bson.M{
		"$inc": bson.M{
			"contagem_global": 1,
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}
