package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"email-job/internal/config"
	"email-job/internal/emails"
	"email-job/internal/infrastructures/client/mailtrap"
	"email-job/internal/infrastructures/rabbitmq"
	"email-job/internal/infrastructures/redis"
	"email-job/internal/infrastructures/sqlite"
	"email-job/internal/middleware"
	"email-job/internal/routes"
	"email-job/pkg"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg         := config.NewConfig()
	validate    := pkg.NewValidator()
	redisClient := redis.NewRedisClient(cfg.REDISaddr, cfg.REDISpassword)
	db, _       := sqlite.NewSQLiteDB(cfg.DSN)
	rmq, _      := rabbitmq.NewRabbitMQClient(cfg.RabbitMQ, logger)

	mailtrapClient, err := mailtrap.NewMailTrapClient(cfg)
	if err != nil {
		log.Fatal(err)
	}

	middleware   := middleware.NewAuthMiddleware(cfg.AdminSecret)

	emailRepo    := emails.NewEmailRepository(db)
	emailSvc     := emails.NewEmailService(mailtrapClient, redisClient, rmq.Channel(), cfg.MailOwner, emailRepo, validate)
	emailHandler := emails.NewEmailHandler(emailSvc)

	r := routes.NewUserRouter(emailHandler, middleware)

	srv := &http.Server{
		Addr:         ":" + cfg.APPport,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		log.Printf("server running on :%s", cfg.APPport)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	go func() {
		ctx := context.Background()
		if err := rmq.StartWorker(ctx, emailSvc); err != nil {
			logger.Error("rabbitmq: worker stopped with error", slog.Any("error", err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	sqlDB, err := db.DB()
	if err == nil {
		if err := sqlDB.Close(); err != nil {
			log.Printf("sqlite: close error %v", err)
		}
	}

	if err := redisClient.Close(); err != nil {
		log.Printf("redis: close error: %v", err)
	}

	log.Println("server exited")
}