package dep

import (
	"context"

	"layout/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	_ "github.com/jackc/pgx/v4/stdlib"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gtracing "gorm.io/plugin/opentelemetry/tracing"
)

type Gorm struct {
	// TODO wrapped database client
	DB     *gorm.DB
	Logger log.Logger
}

func openDB(ctx context.Context, c *conf.Data, tp trace.TracerProvider) (*gorm.DB, error) {
	ctx, span := tp.Tracer("gorm").Start(ctx, "openDB")
	defer span.End()
	span.SetAttributes(attribute.String("driver", c.Database.Driver))
	span.SetAttributes(attribute.String("source", c.Database.Source))
	retries := -1
	var err error = nil
	var db *gorm.DB = nil
	for {
		retries++
		if retries > 3 {
			if err != nil {
				break
			}
			return nil, err
		}
		span.SetAttributes(attribute.Int("retry", retries))
		db, err = gorm.Open(
			postgres.New(
				postgres.Config{
					DriverName:           c.Database.Driver,
					DSN:                  c.Database.Source,
					PreferSimpleProtocol: true,
				},
			),
			&gorm.Config{},
		)
		if err != nil {
			continue
		}
		if db == nil {
			continue
		}
		err := db.Use(gtracing.NewPlugin(gtracing.WithTracerProvider(tp)))
		if err != nil {
			continue
		}
		return db, nil
	}
	return nil, err
}

func NewGorm(c *conf.Data, logger log.Logger, tp trace.TracerProvider) (*Gorm, func(), error) {
	db, err := openDB(context.Background(), c, tp)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		log.NewHelper(logger).Info("closing gorm data resources")
	}

	return &Gorm{
		DB:     db,
		Logger: logger,
	}, cleanup, nil
}

func GormMigrate(ctx context.Context, c *conf.Data, logger log.Logger, models ...interface{}) {
	ctx, span := otel.Tracer("data").Start(ctx, "Migrate")
	defer span.End()
	l := log.NewHelper(logger)
	retries := -1
	for {
		retries++
		span.SetAttributes(attribute.Int("retry", retries))
		if retries > 3 {
			l.Errorf("failed to migrate the schema, after %d", retries)
			return
		}
		l.Infof("migrating the schema, retry: %d", retries)
		client, err := openDB(ctx, c, otel.GetTracerProvider())
		if err != nil {
			l.Errorf("failed opening database: %s", err)
			continue
		}
		for _, model := range models {
			migrator := client.Migrator()
			err = migrator.AutoMigrate(model)
			if err != nil {
				l.Errorf("failed migrating the schema: %s", err)
				continue
			} else {
				l.Infof("migrated the schema successfully")
			}
			break
		}
		if err != nil {
			l.Errorf("failed migrating the schema: %s", err)
			continue
		}
		break
	}
}
