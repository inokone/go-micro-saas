package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cskr/pubsub/v2"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	docs "github.com/inokone/go-micro-saas/api"
	"github.com/inokone/go-micro-saas/internal/auth/account"
	"github.com/inokone/go-micro-saas/internal/auth/role"
	"github.com/inokone/go-micro-saas/internal/auth/user"
	"github.com/inokone/go-micro-saas/internal/common"
	"github.com/inokone/go-micro-saas/internal/history"
	"github.com/inokone/go-micro-saas/internal/mail"
	"github.com/inokone/go-micro-saas/internal/notification"
	"github.com/inokone/go-micro-saas/internal/routes"
)

var (
	Config  *common.AppConfig
	storers routes.Storers
	DB      *sqlx.DB
)

func initStorers() {
	storers.Roles = role.NewPostgresStorer(DB)
	storers.Users = user.NewPostgresStorer(DB, storers.Roles)
	storers.Accounts = account.NewPostgresStorer(DB)
	storers.History = history.NewPostgresStorer(DB)
}

func App(c *common.AppConfig) {
	Config = c
	initStorers()

	// Listen OS for signals - graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	listenOS(cancel)

	ps := pubsub.New[string, common.Event](0)

	startHistoryService(ctx, ps)

	startNotificationService(ctx, ps)

	startGin(ps)
}

func startNotificationService(ctx context.Context, ps *pubsub.PubSub[string, common.Event]) {
	ch := ps.Sub(common.NotificationTopic)
	s := notification.NewService(ch, mail.NewService(Config.Mail, ps))
	s.Start(ctx)
}

func startHistoryService(ctx context.Context, ps *pubsub.PubSub[string, common.Event]) {
	ch := ps.Sub(common.HistoryTopic)
	s := history.NewService(ch, history.NewPostgresStorer(DB))
	s.Start(ctx)
}

func startGin(ps *pubsub.PubSub[string, common.Event]) {
	router := createRouter(ps)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", Config.Web.Port),
		Handler: router.Handler(),
	}

	log.Info("The application is accepting connections...")
	go func() {
		if len(Config.Auth.TLSCert) > 0 {
			if err := srv.ListenAndServeTLS(Config.Auth.TLSCert, Config.Auth.TLSKey); err != nil && err != http.ErrServerClosed {
				log.WithError(err).Error("Failed to initialize the application")
			}
		} else {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.WithError(err).Error("Failed to initialize the application")
			}
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	// catching ctx.Done(). timeout of 3 seconds.
	<-ctx.Done()
	log.Info("The application successfully shut down.")
}

func createRouter(ps *pubsub.PubSub[string, common.Event]) *gin.Engine {
	router := gin.New()
	if Config.Log.PrettyLog {
		router.Use(gin.Logger())
	} else {
		router.Use(common.LoggerMiddleware(log.StandardLogger()))
	}
	router.Use(gin.Recovery())

	docs.SwaggerInfo.BasePath = "/api/v1"
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	return router
}
