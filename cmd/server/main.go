// Command server is the single composition root for goticket. All concrete
// dependencies are wired here; domain packages depend only on interfaces.
package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/quangdangfit/goticket/config"
	"github.com/quangdangfit/goticket/internal/dbs"
	"github.com/quangdangfit/goticket/internal/server"
	"github.com/quangdangfit/goticket/pkg/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	logger.Init(cfg.Log.Level)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	mysql, err := dbs.NewMySQL(cfg.MySQL)
	if err != nil {
		slog.Error("init mysql", "err", err)
	}
	if mysql != nil {
		defer func() { _ = mysql.Close() }()
	}

	rdb, err := dbs.NewRedis(ctx, cfg.Redis)
	if err != nil {
		slog.Error("init redis", "err", err)
	}
	if rdb != nil {
		defer func() { _ = rdb.Close() }()
	}

	pub := dbs.NewKafkaPublisher(cfg.Kafka)
	defer func() { _ = pub.Close() }()

	srv := server.New(cfg.App)
	// Domain packages will be registered here in later phases:
	//   user.Register(srv.APIGroup(), userHandler)
	//   event.Register(srv.APIGroup(), eventHandler)
	//   ...

	if err := srv.Run(ctx); err != nil {
		slog.Error("server stopped", "err", err)
	}
}
