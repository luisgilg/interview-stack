package nosql

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/example/interview-stack/go-service/internal/domain"
)

// ProductStore persists products within MongoDB.
type ProductStore struct {
	collection       *mongo.Collection
	operationTimeout time.Duration
}

type productDocument struct {
	ID        string    `bson:"_id"`
	Name      string    `bson:"name"`
	Price     float64   `bson:"price"`
	Tags      []string  `bson:"tags"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

func NewProductStore(collection *mongo.Collection, operationTimeout time.Duration) *ProductStore {
	return &ProductStore{
		collection:       collection,
		operationTimeout: operationTimeout,
	}
}

func (s *ProductStore) ListProducts(ctx context.Context) ([]domain.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, s.operationTimeout)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := s.collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var products []domain.Product
	for cursor.Next(ctx) {
		var doc productDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		products = append(products, toDomain(doc))
	}
	return products, cursor.Err()
}

func (s *ProductStore) GetProduct(ctx context.Context, id string) (*domain.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, s.operationTimeout)
	defer cancel()

	var doc productDocument
	err := s.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	product := toDomain(doc)
	return &product, nil
}

func (s *ProductStore) CreateProduct(ctx context.Context, product domain.Product) (*domain.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, s.operationTimeout)
	defer cancel()

	if product.ID == "" {
		product.ID = uuid.NewString()
	}
	if product.Tags == nil {
		product.Tags = []string{}
	}

	if _, err := s.collection.InsertOne(ctx, toDocument(product)); err != nil {
		return nil, err
	}
	return &product, nil
}

func (s *ProductStore) UpdateProduct(ctx context.Context, id string, product domain.Product) (*domain.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, s.operationTimeout)
	defer cancel()

	if product.Tags == nil {
		product.Tags = []string{}
	}
	product.ID = id

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	update := bson.M{
		"$set": bson.M{
			"name":       product.Name,
			"price":      product.Price,
			"tags":       product.Tags,
			"updated_at": product.UpdatedAt,
		},
	}
	var doc productDocument
	err := s.collection.FindOneAndUpdate(ctx, bson.M{"_id": id}, update, opts).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	result := toDomain(doc)
	return &result, nil
}

func (s *ProductStore) DeleteProduct(ctx context.Context, id string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, s.operationTimeout)
	defer cancel()
	res, err := s.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return false, err
	}
	return res.DeletedCount > 0, nil
}

func (s *ProductStore) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, s.operationTimeout)
	defer cancel()
	return s.collection.Database().Client().Ping(ctx, readpref.Primary())
}

func (s *ProductStore) EnsureIndexes(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, s.operationTimeout)
	defer cancel()
	indexName := "idx_products_name"
	_, err := s.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: &options.IndexOptions{Name: &indexName},
	})
	if mongo.IsDuplicateKeyError(err) {
		return nil
	}
	return err
}

func toDomain(doc productDocument) domain.Product {
	tags := append([]string{}, doc.Tags...)
	if tags == nil {
		tags = []string{}
	}
	return domain.Product{
		ID:        doc.ID,
		Name:      doc.Name,
		Price:     doc.Price,
		Tags:      tags,
		CreatedAt: doc.CreatedAt,
		UpdatedAt: doc.UpdatedAt,
	}
}

func toDocument(product domain.Product) bson.M {
	return bson.M{
		"_id":        product.ID,
		"name":       product.Name,
		"price":      product.Price,
		"tags":       product.Tags,
		"created_at": product.CreatedAt,
		"updated_at": product.UpdatedAt,
	}
}
