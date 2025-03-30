package dep

import (
	"context"

	"layout/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Mongo struct {
	DB     *mongo.Database
	Logger log.Logger
}

func connMongo(c *conf.Data, logger log.Logger, database string) (*mongo.Database, *mongo.Client, error) {
	uri := c.Mongo.GetUri()
	if uri == "" {
		log.NewHelper(logger).Warn("mongodb uri is empty, using default value mongodb://localhost:27017")
		uri = "mongodb://localhost:27017"
	}
	opts := options.Client()
	opts.ApplyURI(uri)

	username := c.Mongo.GetUsername()
	password := c.Mongo.GetPassword()
	if username != "" && password != "" {
		opts.SetAuth(options.Credential{
			Username: username,
			Password: password,
		})
	}

	client, err := mongo.Connect(opts)
	if err != nil {
		log.NewHelper(logger).Error("failed to connect to mongodb", err)
		return nil, nil, err
	}

	log.NewHelper(logger).Info("pinging mongodb")
	var result bson.M
	if err := client.Database("admin").RunCommand(context.TODO(), bson.D{{"ping", 1}}).Decode(&result); err != nil {
		log.NewHelper(logger).Error("failed to ping mongodb", err)
		panic(err)
	}

	db := client.Database(database)
	return db, client, nil
}
func NewMongo(c *conf.Data, logger log.Logger) (*Mongo, func(), error) {
	log.NewHelper(logger).Info("Initiating NewData")
	database := c.Mongo.GetDatabase()
	db, client, err := connMongo(c, logger, database)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
		retries := 0
		for {
			if retries > 3 {
				return
			}
			err := client.Disconnect(context.Background())
			if err != nil {
				retries++
				log.NewHelper(logger).Error("disconnect mongodb error", err)
			}
			log.NewHelper(logger).Info("disconnected from mongodb successfully")
		}
	}

	log.NewHelper(logger).Info("connected to database successfully")
	return &Mongo{
		DB:     db,
		Logger: logger,
	}, cleanup, nil
}
