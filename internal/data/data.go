package data

import (
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/surrealdb/surrealdb.go"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"

	"layout/ent"
	"layout/internal/conf"
	"layout/internal/dep"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is data providers.
var DataProviderSet = wire.NewSet(NewData, NewUsersRepo, NewProductsRepo)

// Data .
type Data struct {
	gorm    *gorm.DB
	mongo   *mongo.Database
	surreal *surrealdb.DB
	ent     *ent.Client
	logger  log.Logger
}

// NewData .
func NewData(c *conf.Data, logger log.Logger, tp trace.TracerProvider) (*Data, func(), error) {
	lg := log.NewHelper(logger)
	var g *dep.Gorm
	var m *dep.Mongo
	var s *dep.Surreal
	var e *dep.Ent
	var mongoClean func()
	var surrealClean func()
	var entClean func()
	var err error
	noDB := true

	if c.GetDatabase() != nil && c.GetDatabase().Source != "" && c.Database.GetDriver() != "" {
		// g, _, err = dep.NewGorm(c, logger, tp)
		// if err != nil {
		// 	lg.Warn("failed to connect to PostgreSQL", err)
		// 	return nil, nil, err
		// }
		// noDB = false
		//
		c, cleanup, err := dep.NewEnt(c, logger, tp.Tracer("ent"))
		if err != nil {
			lg.Warn("failed to connect to ent", err)
		}
		entClean = cleanup
		e = c
		noDB = false
	}

	if c.GetMongo != nil && c.GetMongo().GetUri() != "" && c.GetMongo().GetDatabase() != "" && c.GetMongo().GetUsername() != "" && c.GetMongo().GetPassword() != "" {
		m, mongoClean, err = dep.NewMongo(c, logger)
		if err != nil {
			lg.Warn("failed to connect to MongoDB", err)
			return nil, nil, err
		}
		noDB = false
	}

	if c.GetSurreal() != nil && c.GetSurreal().GetAddr() != "" && c.GetSurreal().GetDatabase() != "" && c.GetSurreal().GetNamespace() != "" && c.GetSurreal().GetUsername() != "" && c.GetSurreal().GetPassword() != "" {
		s, surrealClean, err = dep.NewSurreal(c, logger)
		if err != nil {
			lg.Warn("failed to connect to SurrealDB", err)
			return nil, nil, err
		}
		noDB = false
	}

	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
		if mongoClean != nil {
			mongoClean()
		}
		if surrealClean != nil {
			surrealClean()
		}
		if entClean != nil {
			entClean()
		}
	}

	if noDB {
		return nil, nil, errors.InternalServer("no database configured", "no database configured")
	}

	data := &Data{logger: logger}
	if g != nil {
		lg.Debug("Attaching PostgreSQL")
		data.gorm = g.DB
	}
	if m != nil {
		lg.Debug("Attaching MongoDB")
		data.mongo = m.DB
	}
	if s != nil {
		lg.Debug("Attaching SurrealDB")
		data.surreal = s.DB
	}
	if e != nil {
		lg.Debug("Attaching ent")
		data.ent = e.Client
	}

	return data, cleanup, nil
}
