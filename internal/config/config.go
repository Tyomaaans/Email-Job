package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	APPport        string
	DSN            string
	REDISaddr      string
	REDISpassword  string
	RabbitMQ       string
	MailTrapApiKey string
	MailTrapName   string
	MailTrapRole   string
	MailOwner      string
	AdminSecret    string
}

func NewConfig() AppConfig {
	err := godotenv.Load()
	if err != nil {
		log.Println(".env file not found!")
	}

	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
        log.Fatal("APP_PORT environment variable is required!")
    }

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environtment variable is required!")
	}

	redisAddr := os.Getenv("REDIS_ADDR")
    if redisAddr == "" {
        log.Fatal("REDIS_ADDR environment variable is required!")
    }

	redisPassword := os.Getenv("REDIS_PASSWORD")
	if redisPassword == "" {
        log.Fatal("REDIS_PASSWORD environment variable is required!")
    }

	rabbitMQ := os.Getenv("RABBITMQ_DSN")
	if rabbitMQ == "" {
		log.Fatal("RABBITMQ_DSN environment variable is required!")

	}

	mailTrapApiKey := os.Getenv("MAILTRAP_API_KEY")
	if mailTrapApiKey == "" {
		log.Fatal("MAILTRAP_API_KEY environment variable is required!")
	}

	mailTrapName := os.Getenv("MAILTRAP_NAME")
	if mailTrapName == "" {
		log.Fatal("MAILTRAP_Name environment variable is required!")
	}

	mailTrapRole := os.Getenv("MAILTRAP_ROLE")
	if mailTrapRole == "" {
		log.Fatal("MAILTRAP_ROLE environment variable is required!")
	}

	mailOwner := os.Getenv("MAIL_OWNER")
	if mailOwner == "" {
		log.Fatal("MAIL_OWNER environment variable is required!")
	}

	adminSecret := os.Getenv("ADMIN_SECRET_KEY")
	if adminSecret == "" {
		log.Fatal("ADMIN_SECRET_KEY environment variable is required!")
	}

	return AppConfig{
		APPport:        appPort,
		DSN:            dsn,
		REDISaddr:      redisAddr,
		REDISpassword:  redisPassword,
		RabbitMQ:       rabbitMQ,
		MailTrapApiKey: mailTrapApiKey,
		MailTrapName:   mailTrapName,
		MailTrapRole:   mailTrapRole,
		MailOwner:      mailOwner,
		AdminSecret:    adminSecret,
	}
}