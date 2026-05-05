// Command server is the single composition root for goticket. All concrete
// dependencies are wired here; domain packages depend only on interfaces.
package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

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
	promohttp "github.com/quangdangfit/goticket/internal/promo/port/http"
	promorepo "github.com/quangdangfit/goticket/internal/promo/repository"
	promosvc "github.com/quangdangfit/goticket/internal/promo/service"
	notifsender "github.com/quangdangfit/goticket/internal/notification/sender"
	notifsvc "github.com/quangdangfit/goticket/internal/notification/service"
	"github.com/quangdangfit/goticket/internal/server"
	"github.com/quangdangfit/goticket/internal/server/middleware"
	userdto "github.com/quangdangfit/goticket/internal/user/dto"
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

// userEmailLookup adapts user.Service to notification's Profile port.
type userEmailLookup struct {
	svc interface {
		Profile(ctx context.Context, userID string) (*userdto.Profile, error)
	}
}

func (u userEmailLookup) Email(ctx context.Context, userID string) (string, error) {
	p, err := u.svc.Profile(ctx, userID)
	if err != nil {
		return "", err
	}
	return p.Email, nil
}

// jwtVerifier adapts jwt.Manager to middleware.TokenVerifier.
type jwtVerifier struct{ m jwt.Manager }

func (v jwtVerifier) Verify(t string) (string, string, error) { return v.m.Verify(t) }

// invSvcAlias bundles the two inventory ports we need at the wiring site
// (inventory.Inventory for /holds, ticket.AvailabilityReader for ticket
// listing). Concrete redisInventory satisfies both.
type invSvcAlias = inventory.Inventory

// promoApplier adapts promo.Service to order.PromoApplier (only Apply
// is needed; Redeem is invoked separately on payment success).
type promoApplier struct {
	svc interface {
		Apply(ctx context.Context, code, userID string, subtotal int64) (int64, error)
	}
}

func (p promoApplier) Apply(ctx context.Context, code, userID string, subtotal int64) (int64, error) {
	return p.svc.Apply(ctx, code, userID, subtotal)
}

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

			pmRepo := promorepo.New(mysql)
			pmSvc := promosvc.New(pmRepo)
			promohttp.RegisterRoutes(srv.APIGroup(), promohttp.NewHandler(pmSvc), verifier)

			oSvc := ordersvc.New(oRepo, inv, tSvc, promoApplier{pmSvc}, idemGuard)
			ordersRL := middleware.RateLimit(rdb.Client(), cfg.RateLimit.OrdersPerMin, func(c *gin.Context) string {
				if uid, ok := c.Get("user_id"); ok {
					if s, _ := uid.(string); s != "" {
						return "u:" + s
					}
				}
				return "ip:" + c.ClientIP()
			})
			orderhttp.RegisterRoutes(srv.APIGroup(), orderhttp.NewHandler(oSvc), verifier, ordersRL)

			pRepo := payrepo.New(mysql)
			pSvc := paysvc.New(pRepo, payprovmock.New(), oSvc, pub)
			payhttp.RegisterRoutes(srv.WebhookGroup(), payhttp.NewHandler(pSvc))

			if len(cfg.Kafka.Brokers) > 0 {
				cons := notifsvc.New(cfg.Kafka.Brokers, "goticket-notifications",
					notifsender.NewLog(), userEmailLookup{uSvc})
				go func() {
					if err := cons.Run(ctx); err != nil {
						slog.Error("notification consumer", "err", err)
					}
				}()
			}
		}
	}

	if err := srv.Run(ctx); err != nil {
		slog.Error("server stopped", "err", err)
	}
}

