package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/ssddi456/sms_server/internal/config"
	"github.com/ssddi456/sms_server/internal/handler"
	"github.com/ssddi456/sms_server/internal/logging"
	"github.com/ssddi456/sms_server/internal/middleware"
	"github.com/ssddi456/sms_server/internal/store"
)

func main() {
	cfg := config.Load()
	gin.SetMode(cfg.GinMode)
	logger := logging.NewLogger(cfg.LogLevel)

	storeInstance, err := store.NewStore(cfg.DataDir)
	if err != nil {
		log.Fatalf("init store: %v", err)
	}
	defer storeInstance.Close()

	h := handler.New(storeInstance, logger)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.RequestLog(logger))

	smsGroup := r.Group("/sms")
	smsGroup.GET("/healthz", h.Health)
	smsGroup.POST("/webhook", h.Webhook)
	smsGroup.GET("/history", h.History)

	legacyGroup := r.Group("/v1/sms")
	legacyGroup.GET("/healthz", h.Health)
	legacyGroup.POST("/receive", h.Webhook)
	legacyGroup.GET("/history", h.History)

	logger.Info("server_start", map[string]interface{}{
		"host": cfg.Host,
		"port": cfg.Port,
		"data_dir": cfg.DataDir,
	})
	fmt.Printf("listening on %s\n", cfg.ListenAddr)
	if err := r.Run(cfg.ListenAddr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
