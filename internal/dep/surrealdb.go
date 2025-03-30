package dep

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/surrealdb/surrealdb.go"
	"layout/internal/conf"
)

type Surreal struct {
	DB  *surrealdb.DB
	Log *log.Helper
}

func NewSurreal(c *conf.Data, logger log.Logger) (*Surreal, func(), error) {
	lg := log.NewHelper(logger)
	lg.Debug("Initiating NewSurreal")
	connURL := c.GetSurreal().GetAddr()
	db, err := surrealdb.New(connURL)
	if err != nil {
		return nil, nil, err
	}

	namespace := c.GetSurreal().GetNamespace()
	database := c.GetSurreal().GetDatabase()

	lg.Debug("connecting to surreal")
	if err = db.Use(namespace, database); err != nil {
		return nil, nil, err
	}

	authData := &surrealdb.Auth{
		Username: c.GetSurreal().GetUsername(),
		Password: c.GetSurreal().GetPassword(),
	}
	lg.Debug("signing in to surreal")
	token, err := db.SignIn(authData)
	if err != nil {
		return nil, nil, err
	}
	lg.Debug("authenticating to surreal")
	if err := db.Authenticate(token); err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		lg.Debug("closing the surreal connection")
		if err := db.Invalidate(); err != nil {
			lg.Error("surreal invalidate error", err)
		}
		err := db.Close()
		if err != nil {
			lg.Error("surreal close error", err)
		}
	}

	lg.Debug("connected to surreal successfully")
	return &Surreal{
		DB:  db,
		Log: lg,
	}, cleanup, nil
}
