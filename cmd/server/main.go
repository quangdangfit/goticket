// Command server is the single composition root for goticket. All concrete
// dependencies are wired here; domain packages depend only on interfaces.
package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/quangdangfit/goticket/config"
	"github.com/quangdangfit/goticket/internal/dbs"
	"github.com/quangdangfit/goticket/internal/event"
	"github.com/quangdangfit/goticket/internal/inventory"
	eventhttp "github.com/quangdangfit/goticket/internal/event/port/http"
	eventrepo "github.com/quangdangfit/goticket/internal/event/repository"
	eventsvc "github.com/quangdangfit/goticket/internal/event/service"
	invhttp "github.com/quangdangfit/goticket/internal/inventory/port/http"
	invsvc "github.com/quangdangfit/goticket/internal/inventory/service"
	orderhttp "github.com/quangdangfit/goticket/internal/order/port/http"
	orderrepo "github.com/quangdangfit/goticket/internal/order/repository"
	ordersvc "github.com/quangdangfit/goticket/internal/order/service"
	payhttp "github.com/quangdangfit/goticket/internal/payment/port/http"
	payprovmock "github.com/quangdangfit/goticket/internal/payment/provider/mock"
	payrepo "github.com/quangdangfit/goticket/internal/payment/repository"
	paysvc "github.com/quangdangfit/goticket/internal/payment/service"
	"github.com/quangdangfit/goticket/internal/server"
	tickethttp "github.com/quangdangfit/goticket/internal/ticket/port/http"
	ticketrepo "github.com/quangdangfit/goticket/internal/ticket/repository"
	ticketsvc "github.com/quangdangfit/goticket/internal/ticket/service"
	userhttp "github.com/quangdangfit/goticket/internal/user/port/http"
	userrepo "github.com/quangdangfit/goticket/internal/user/repository"
	usersvc "github.com/quangdangfit/goticket/internal/user/service"
	"github.com/quangdangfit/goticket/pkg/idempotency"
	"github.com/quangdangfit/goticket/pkg/jwt"
	"github.com/quangdangfit/goticket/pkg/logger"
)

// jwtVerifier adapts jwt.Manager to middleware.TokenVerifier.
type jwtVerifier struct{ m jwt.Manager }

func (v jwtVerifier) Verify(t string) (string, string, error) { return v.m.Verify(t) }

// invSvcAlias bundles the two inventory ports we need at the wiring site
// (inventory.Inventory for /holds, ticket.AvailabilityReader for ticket
// listing). Concrete redisInventory satisfies both.
type invSvcAlias = inventory.Inventory

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

		var ec event.Cache
		if rdb != nil {
			ec = eventrepo.NewCache(rdb.Client(), 30*time.Second)
		}
		eRepo := eventrepo.New(mysql)
		eSvc := eventsvc.New(eRepo, ec)
		eventhttp.RegisterRoutes(srv.APIGroup(), eventhttp.NewHandler(eSvc), verifier)

		tRepo := ticketrepo.New(mysql)
		var inv = func() interface {
			invSvcAlias
		} {
			if rdb == nil {
				return nil
			}
			return invsvc.New(rdb.Client())
		}()
		tSvc := ticketsvc.New(tRepo, inv)
		tickethttp.RegisterRoutes(srv.APIGroup(), tickethttp.NewHandler(tSvc), verifier)
		if inv != nil {
			invhttp.RegisterRoutes(srv.APIGroup(), invhttp.NewHandler(inv), verifier)
		}

		if inv != nil && rdb != nil {
			oRepo := orderrepo.New(mysql)
			idemGuard := idempotency.New(rdb.Client())
			oSvc := ordersvc.New(oRepo, inv, tSvc, nil, idemGuard)
			orderhttp.RegisterRoutes(srv.APIGroup(), orderhttp.NewHandler(oSvc), verifier, nil)

			pRepo := payrepo.New(mysql)
			pSvc := paysvc.New(pRepo, payprovmock.New(), oSvc, pub)
			payhttp.RegisterRoutes(srv.WebhookGroup(), payhttp.NewHandler(pSvc))
		}
	}

	if err := srv.Run(ctx); err != nil {
		slog.Error("server stopped", "err", err)
	}
}

