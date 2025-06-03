package gogi

import (
	"context"
	"fmt"
	"log"
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

	if readPref != "" {
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
			log.Printf("[Mongo] Unknown read preference: %s, defaulting to primary", readPref)
			opts.SetReadPreference(readpref.Primary())
		}
	}

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	log.Printf("[Mongo] Connected to %s", cfg.URI)
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

// ---------- Helpers ----------

func (c *MongoClient) FindOne(
	ctx context.Context,
	db string,
	coll string,
	filter any,
	dest any,
	opts ...*options.FindOneOptions,
) error {
	message := fmt.Sprintf("[Mongo] FindOne: %s.%s | filter=%v", db, coll, filter)
	GetLogger().Debug(message)

	err := c.Reader.Database(db).Collection(coll).
		FindOne(ctx, filter, opts...).
		Decode(dest)
	if err == mongo.ErrNoDocuments {
		return nil // match MySQL "no rows" behavior
	}
	return err
}

func (c *MongoClient) FindMany(
	ctx context.Context,
	db string,
	coll string,
	filter any,
	handle func(*mongo.Cursor) error,
	opts ...*options.FindOptions,
) error {
	message := fmt.Sprintf("[Mongo] FindMany: %s.%s | filter=%v", db, coll, filter)
	GetLogger().Debug(message)

	cursor, err := c.Reader.Database(db).Collection(coll).
		Find(ctx, filter, opts...)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	return handle(cursor)
}

func (c *MongoClient) InsertOne(
	ctx context.Context,
	db string,
	coll string,
	doc any,
	opts ...*options.InsertOneOptions,
) (*mongo.InsertOneResult, error) {
	message := fmt.Sprintf("[Mongo] InsertOne: %s.%s", db, coll)
	GetLogger().Debug(message)

	return c.Writer.Database(db).Collection(coll).InsertOne(ctx, doc, opts...)
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
