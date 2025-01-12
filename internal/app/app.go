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
	"github.com/inokone/go-micro-saas/db"
	"github.com/inokone/go-micro-saas/internal/common"
)

var (
	Config  *common.AppConfig
	storers routes.Storers
	DB      *sqlx.DB
)

func initApp(path string) {
	var err error
	Config, err = common.LoadConfig(path)
	if err != nil {
		log.WithError(err).Error("Failed to load application configuration.")
		os.Exit(1)
	}
	common.InitLogging(Config.Log)
	DB, err = db.InitDB(Config.DB)
	if err != nil {
		log.WithError(err).Error("Failed to connect to database.")
		os.Exit(1)
	}
	initStorers()
}

func initStorers() {
	storers.Roles = role.NewPostgresStorer(DB)
	storers.Users = user.NewPostgresStorer(DB, storers.Roles)
	storers.Accounts = account.NewPostgresStorer(DB)
	storers.History = history.NewPostgresStorer(DB)
}

func App(path string) {
	initApp(path)

	startGin(ps)
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
