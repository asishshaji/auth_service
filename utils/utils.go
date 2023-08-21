package utils

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type EnvironmentConfig struct {
	APP_PORT    string
	DB_HOST     string
	DB_PORT     string
	DB_NAME     string
	DB_USER     string
	DB_PASSWORD string
	SSLMODE     string
	l           *log.Logger
}

func LoadEnv(l *log.Logger) *EnvironmentConfig {
	if err := godotenv.Load(); err != nil {
		l.Fatalln("Error loading env file")
	}

	return &EnvironmentConfig{
		APP_PORT:    os.Getenv("APP_PORT"),
		DB_PORT:     os.Getenv("DB_PORT"),
		DB_HOST:     os.Getenv("DB_HOST"),
		DB_NAME:     os.Getenv("DB_NAME"),
		DB_USER:     os.Getenv("DB_USER"),
		SSLMODE:     os.Getenv("SSL_MODE"),
		DB_PASSWORD: os.Getenv("DB_PASSWORD"),
		l:           l,
	}
}

func (c *EnvironmentConfig) InitDB() *sqlx.DB {
	db, err := sqlx.Connect("postgres",
		fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", c.DB_HOST, c.DB_PORT, c.DB_USER, c.DB_PASSWORD, c.DB_NAME, c.SSLMODE))
	if err != nil {
		c.l.Fatalf("error connecting to DB: %v\n", err)
	}

	// db.SetMaxOpenConns()
	// db.SetMaxIdleConns(c.MaxIdle)
	// db.SetConnMaxLifetime(c.MaxLifetime)

	err = installSchema(db)
	if err != nil {
		c.l.Fatalf("error creating schema %s\n", err)
	}
	return db
}

func installSchema(db *sqlx.DB) error {
	file, err := os.Open("schema.sql")
	if err != nil {
		return err
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	if _, err := db.Exec(string(data)); err != nil {
		return err
	}

	return nil

}
