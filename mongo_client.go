package gogi

import (
	"context"
	"fmt"
	"time"

	"github.com/dejaniskra/go-gi/internal/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const defaultMongoTimeout = 10 * time.Second

type MongoClient struct {
	Writer *mongo.Client
	Reader *mongo.Client
}

var mongoClients = make(map[string]*MongoClient)

func GetMongoClient(role string) (*MongoClient, error) {
	if client, exists := mongoClients[role]; exists {
		return client, nil
	}

	cfg := config.GetConfig().Mongo[role]
	if cfg == nil {
		return nil, fmt.Errorf("no MongoDB configuration found for role: %s", role)
	}

	client, err := newMongoClient(cfg)
	if err != nil {
		return nil, err
	}

	mongoClients[role] = client
	return client, nil
}

func newMongoClient(cfg *config.MongoRoleConfig) (*MongoClient, error) {
	writer, err := newMongoConnection(cfg.Writer, "primary")
	if err != nil {
		return nil, fmt.Errorf("failed to create writer connection: %w", err)
	}

	var reader *mongo.Client
	if cfg.Reader != nil {
		reader, err = newMongoConnection(cfg.Reader, cfg.Reader.ReadPreference)
		if err != nil {
			return nil, fmt.Errorf("failed to create reader connection: %w", err)
		}
	} else {
		reader = writer
	}

	return &MongoClient{
		Writer: writer,
		Reader: reader,
	}, nil
}

func newMongoConnection(cfg *config.MongoConfig, readPref string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultMongoTimeout)
	defer cancel()

	opts := options.Client().ApplyURI(cfg.URI)

	switch readPref {
	case "primary":
		opts.SetReadPreference(readpref.Primary())
	case "primaryPreferred":
		opts.SetReadPreference(readpref.PrimaryPreferred())
	case "secondary":
		opts.SetReadPreference(readpref.Secondary())
	case "secondaryPreferred":
		opts.SetReadPreference(readpref.SecondaryPreferred())
	case "nearest":
		opts.SetReadPreference(readpref.Nearest())
	default:
		GetLogger().Debug(fmt.Sprintf("[Mongo] Unknown read preference: %s, defaulting to primary", readPref))
		opts.SetReadPreference(readpref.Primary())
	}

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	GetLogger().Debug(fmt.Sprintf("[Mongo] Connected to %s", cfg.URI))
	return client, nil
}

func (c *MongoClient) Ping(ctx context.Context) error {
	if err := c.Writer.Ping(ctx, nil); err != nil {
		return fmt.Errorf("writer ping failed: %w", err)
	}
	if c.Reader != c.Writer {
		if err := c.Reader.Ping(ctx, nil); err != nil {
			return fmt.Errorf("reader ping failed: %w", err)
		}
	}
	return nil
}

func (c *MongoClient) Close(ctx context.Context) error {
	if err := c.Writer.Disconnect(ctx); err != nil {
		return err
	}
	if c.Reader != c.Writer {
		return c.Reader.Disconnect(ctx)
	}
	return nil
}

// ---------- Query Helpers ----------

func (c *MongoClient) FindOne(
	ctx context.Context,
	db, coll string,
	filter any,
	dest any,
	opts ...*options.FindOneOptions,
) error {
	GetLogger().Debug(fmt.Sprintf("[Mongo] FindOne: %s.%s | filter=%v", db, coll, filter))
	err := c.Reader.Database(db).Collection(coll).FindOne(ctx, filter, opts...).Decode(dest)
	if err == mongo.ErrNoDocuments {
		return nil
	}
	return err
}

func (c *MongoClient) FindMany(
	ctx context.Context,
	db, coll string,
	filter any,
	handle func(*mongo.Cursor) error,
	opts ...*options.FindOptions,
) error {
	GetLogger().Debug(fmt.Sprintf("[Mongo] FindMany: %s.%s | filter=%v", db, coll, filter))
	cur, err := c.Reader.Database(db).Collection(coll).Find(ctx, filter, opts...)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)
	return handle(cur)
}

func (c *MongoClient) InsertOne(
	ctx context.Context,
	db, coll string,
	doc any,
	opts ...*options.InsertOneOptions,
) (*mongo.InsertOneResult, error) {
	GetLogger().Debug(fmt.Sprintf("[Mongo] InsertOne: %s.%s", db, coll))
	return c.Writer.Database(db).Collection(coll).InsertOne(ctx, doc, opts...)
}

func (c *MongoClient) InsertMany(
	ctx context.Context,
	db, coll string,
	docs []any,
	opts ...*options.InsertManyOptions,
) (*mongo.InsertManyResult, error) {
	GetLogger().Debug(fmt.Sprintf("[Mongo] InsertMany: %s.%s", db, coll))
	return c.Writer.Database(db).Collection(coll).InsertMany(ctx, docs, opts...)
}

func (c *MongoClient) UpdateOne(
	ctx context.Context,
	db, coll string,
	filter, update any,
	opts ...*options.UpdateOptions,
) (*mongo.UpdateResult, error) {
	GetLogger().Debug(fmt.Sprintf("[Mongo] UpdateOne: %s.%s | filter=%v", db, coll, filter))
	return c.Writer.Database(db).Collection(coll).UpdateOne(ctx, filter, update, opts...)
}

func (c *MongoClient) UpdateMany(
	ctx context.Context,
	db, coll string,
	filter, update any,
	opts ...*options.UpdateOptions,
) (*mongo.UpdateResult, error) {
	GetLogger().Debug(fmt.Sprintf("[Mongo] UpdateMany: %s.%s | filter=%v", db, coll, filter))
	return c.Writer.Database(db).Collection(coll).UpdateMany(ctx, filter, update, opts...)
}

func (c *MongoClient) DeleteOne(
	ctx context.Context,
	db, coll string,
	filter any,
	opts ...*options.DeleteOptions,
) (*mongo.DeleteResult, error) {
	GetLogger().Debug(fmt.Sprintf("[Mongo] DeleteOne: %s.%s | filter=%v", db, coll, filter))
	return c.Writer.Database(db).Collection(coll).DeleteOne(ctx, filter, opts...)
}

func (c *MongoClient) DeleteMany(
	ctx context.Context,
	db, coll string,
	filter any,
	opts ...*options.DeleteOptions,
) (*mongo.DeleteResult, error) {
	GetLogger().Debug(fmt.Sprintf("[Mongo] DeleteMany: %s.%s | filter=%v", db, coll, filter))
	return c.Writer.Database(db).Collection(coll).DeleteMany(ctx, filter, opts...)
}

func (c *MongoClient) WithTransaction(
	ctx context.Context,
	fn func(mongo.SessionContext) error,
) error {
	session, err := c.Writer.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	return mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		if err := session.StartTransaction(); err != nil {
			return err
		}
		if err := fn(sc); err != nil {
			_ = session.AbortTransaction(sc)
			return err
		}
		return session.CommitTransaction(sc)
	})
}
