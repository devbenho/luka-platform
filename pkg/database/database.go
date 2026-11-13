package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const DatabaseTimeout = 5 * time.Second

//go:generate mockery --name=IDatabase
type IDatabase interface {
	GetDB() *mongo.Database
	WithTransaction(ctx context.Context, function func(sessCtx mongo.SessionContext) error) error
	Transactor
	Create(ctx context.Context, collection string, doc interface{}) error
	CreateInBatches(ctx context.Context, collection string, docs []interface{}) error
	Update(ctx context.Context, collection string, filter, update interface{}) error
	Delete(ctx context.Context, collection string, filter interface{}) error
	DeleteAll(ctx context.Context, collection string, filter interface{}) error
	SoftDelete(ctx context.Context, collection, deletedAtField string, filter interface{}) error
	FindOne(ctx context.Context, collection string, filter, result interface{}) error
	Find(ctx context.Context, collection string, filter, result interface{}) error
	Count(ctx context.Context, collection string, filter interface{}) (int64, error)
}

// Transactor runs fn inside a multi-document transaction: every write made
// with the ctx passed to fn commits together or not at all. fn may be called
// more than once if the transaction is retried after a transient error.
type Transactor interface {
	RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type Database struct {
	client   *mongo.Client
	database *mongo.Database
}

func NewDatabase(uri, dbName string) (*Database, error) {
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), DatabaseTimeout)
	defer cancel()

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	if err := requireReplicaSet(ctx, client); err != nil {
		return nil, err
	}

	database := client.Database(dbName)
	return &Database{
		client:   client,
		database: database,
	}, nil
}

// requireReplicaSet fails fast when the server cannot run transactions
// (a standalone mongod), instead of failing on the first order.
func requireReplicaSet(ctx context.Context, client *mongo.Client) error {
	var hello struct {
		SetName string `bson:"setName"`
		Msg     string `bson:"msg"`
	}
	if err := client.Database("admin").RunCommand(ctx, bson.D{{Key: "hello", Value: 1}}).Decode(&hello); err != nil {
		return fmt.Errorf("checking MongoDB topology: %w", err)
	}
	if hello.SetName == "" && hello.Msg != "isdbgrid" {
		return errors.New("MongoDB must run as a replica set (or sharded cluster) because orders use transactions; " +
			"for local development start mongod with --replSet rs0 and run rs.initiate() (see README)")
	}
	return nil
}

func (d *Database) GetDB() *mongo.Database {
	return d.database
}

func (d *Database) WithTransaction(ctx context.Context, function func(sessCtx mongo.SessionContext) error) error {
	session, err := d.client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		err := function(sessCtx)
		return nil, err
	}

	_, err = session.WithTransaction(ctx, callback)
	return err
}

func (d *Database) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return d.WithTransaction(ctx, func(sessCtx mongo.SessionContext) error {
		return fn(sessCtx)
	})
}

func (d *Database) Create(ctx context.Context, collection string, doc interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, DatabaseTimeout)
	defer cancel()

	_, err := d.database.Collection(collection).InsertOne(ctx, doc)
	return err
}

func (d *Database) CreateInBatches(ctx context.Context, collection string, docs []interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, DatabaseTimeout)
	defer cancel()

	_, err := d.database.Collection(collection).InsertMany(ctx, docs)
	return err
}

func (d *Database) Update(ctx context.Context, collection string, filter, update interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, DatabaseTimeout)
	defer cancel()

	result, err := d.database.Collection(collection).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (d *Database) Delete(ctx context.Context, collection string, filter interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, DatabaseTimeout)
	defer cancel()

	result, err := d.database.Collection(collection).DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (d *Database) DeleteAll(ctx context.Context, collection string, filter interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, DatabaseTimeout)
	defer cancel()

	_, err := d.database.Collection(collection).DeleteMany(ctx, filter)
	return err
}

// SoftDelete sets deletedAtField to the current time on the first document
// matching filter. Each model stores its deletion timestamp under its own
// bson field name, so callers must pass it explicitly.
func (d *Database) SoftDelete(ctx context.Context, collection, deletedAtField string, filter interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, DatabaseTimeout)
	defer cancel()
	update := bson.M{"$set": bson.M{deletedAtField: time.Now()}}
	updateResult, err := d.database.Collection(collection).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if updateResult.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (d *Database) FindOne(ctx context.Context, collection string, filter, result interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, DatabaseTimeout)
	defer cancel()

	err := d.database.Collection(collection).FindOne(ctx, filter).Decode(result)
	return err
}

func (d *Database) Find(ctx context.Context, collection string, filter, result interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, DatabaseTimeout)
	defer cancel()

	cursor, err := d.database.Collection(collection).Find(ctx, filter)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, result)
	return err
}

func (d *Database) Count(ctx context.Context, collection string, filter interface{}) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, DatabaseTimeout)
	defer cancel()

	count, err := d.database.Collection(collection).CountDocuments(ctx, filter)
	return count, err
}
