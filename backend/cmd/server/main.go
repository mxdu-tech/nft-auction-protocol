package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/internal/api"
	"backend/internal/api/middleware"
	"backend/internal/config"
	"backend/internal/indexer"
	"backend/internal/infra/db"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	log.Printf("config loaded: env=%s port=%s chain_id=%d rpc=%s contract=%s",
		cfg.Env, cfg.Port, cfg.ChainID, cfg.RPCURL, cfg.AuctionContract)

	mysqlDB, err := db.NewMySQL(cfg)
	if err != nil {
		log.Fatal("failed to connect database: ", err)
	}
	log.Println("db connected")

	idx, err := indexer.NewIndexer(cfg, mysqlDB)
	if err != nil {
		log.Fatal("failed to init indexer: ", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	idx.Start(ctx)

	r := gin.Default()
	r.Use(middleware.InjectChainContext(cfg))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "env": cfg.Env})
	})

	api.RegisterRoutes(r, mysqlDB)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		log.Println("server starting on :" + cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutdown signal received")

	cancel()

	ctxTimeout, cancelTimeout := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelTimeout()

	if err := srv.Shutdown(ctxTimeout); err != nil {
		log.Fatal("server shutdown failed: ", err)
	}

	log.Println("server exited gracefully")
}
