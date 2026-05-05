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
	userhttp "github.com/quangdangfit/goticket/internal/user/port/http"
	userrepo "github.com/quangdangfit/goticket/internal/user/repository"
	usersvc "github.com/quangdangfit/goticket/internal/user/service"
	"github.com/quangdangfit/goticket/pkg/jwt"
	"github.com/quangdangfit/goticket/pkg/logger"
)

// jwtVerifier adapts jwt.Manager to middleware.TokenVerifier.
type jwtVerifier struct{ m jwt.Manager }

func (v jwtVerifier) Verify(t string) (string, string, error) { return v.m.Verify(t) }

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

	jwtMgr := jwt.New(cfg.JWT.AccessSecret, cfg.JWT.RefreshSecret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)
	verifier := jwtVerifier{m: jwtMgr}

	if mysql != nil {
		uRepo := userrepo.New(mysql)
		uSvc := usersvc.New(uRepo, jwtMgr)
		userhttp.RegisterRoutes(srv.APIGroup(), userhttp.NewHandler(uSvc), verifier)
	}

	if err := srv.Run(ctx); err != nil {
		slog.Error("server stopped", "err", err)
	}
}
