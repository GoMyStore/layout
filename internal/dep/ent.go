package dep

import (
	"layout/ent"
	"layout/internal/conf"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel/trace"
)

type Ent struct {
	Client *ent.Client
	logger *log.Helper
	tp     trace.Tracer
}

func NewEnt(c *conf.Data, logger log.Logger, tracer trace.Tracer) (*Ent, func(), error) {
	cleanup := func() {}
	if c.GetDatabase() == nil || c.GetDatabase().GetSource() == "" {
		return nil, cleanup, errors.InternalServer("missing database configuration", "DB config is missing in config file, please check your config file")
	}
	dsn := c.GetDatabase().GetSource()
	client, err := ent.Open("postgres", dsn)
	if err != nil {
		return nil, cleanup, err
	}
	lg := log.NewHelper(logger)

	cleanup = func() {
		lg.Warn("Closing the ent client")
		client.Close()
	}

	e := &Ent{
		Client: client,
		logger: lg,
		tp:     tracer,
	}
	return e, cleanup, nil
}
