package dep

import (
	"layout/internal/conf"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/trace"
)

type Nats struct {
	Conn   *nats.Conn
	Logger *log.Helper
	tp     trace.Tracer
}

func NewNats(c *conf.Bootstrap, logger log.Logger, tp trace.TracerProvider) (*Nats, error) {
	t := tp.Tracer("nats")
	log := log.NewHelper(logger)

	conn, err := conn(c, log)
	if err != nil {
		return nil, err
	}

	return &Nats{
		Logger: log,
		tp:     t,
		Conn:   conn,
	}, nil
}

func conn(c *conf.Bootstrap, log *log.Helper) (*nats.Conn, error) {
	addr := nats.DefaultURL
	if c.GetData().GetNats().GetAddr() == "" {
		return nil, errors.InternalServer("No nats address provided", "No data.nats.addr was set in the config")
	}
	addr = c.GetData().GetNats().GetAddr()

	name := "unnamed"
	if c.GetMetadata().GetName() != "" {
		name = c.GetMetadata().GetName()
	}
	opts := []nats.Option{
		nats.Name(name),
	}

	nc, err := nats.Connect(addr, opts...)
	if err != nil {
		return nil, err
	}
	defer nc.Close()

	return nc, nil
}
