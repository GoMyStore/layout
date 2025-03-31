package dep

import (
	"context"

	"layout/internal/conf"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.opentelemetry.io/otel/trace"
)

type Nats struct {
	Conn   *nats.Conn
	JS     jetstream.JetStream
	Logger *log.Helper
	tp     trace.Tracer
}

func NewNats(c *conf.Bootstrap, logger log.Logger, tp trace.TracerProvider) (*Nats, error) {
	t := tp.Tracer("nats")
	log := log.NewHelper(logger)
	var conn *nats.Conn

	conn, err := connect(c, log, t, context.Background())
	if err != nil {
		return nil, err
	}

	nats := Nats{
		Logger: log,
		tp:     t,
		Conn:   conn,
	}

	if c.GetData().GetNats().GetJetstream() {
		js, err := nats.connectJS(c)
		if err != nil {
			return nil, err
		}
		return &Nats{
			Logger: log,
			tp:     t,
			Conn:   conn,
			JS:     js,
		}, nil
	}

	return &nats, nil
}

func connect(c *conf.Bootstrap, log *log.Helper, t trace.Tracer, ctx context.Context) (*nats.Conn, error) {
	ctx, span := t.Start(ctx, "dep.Nats.conn")
	defer span.End()
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

func (n *Nats) connectJS(c *conf.Bootstrap) (jetstream.JetStream, error) {
	ctx, span := n.tp.Start(context.Background(), "dep.Nats.connJS")
	defer span.End()

	var conn *nats.Conn = n.Conn
	var err error

	if conn == nil {
		conn, err = connect(c, n.Logger, n.tp, ctx)
		if err != nil {
			return nil, err
		}
	}

	js, err := jetstream.New(conn)
	if err != nil {
		return nil, err
	}

	return js, nil
}
