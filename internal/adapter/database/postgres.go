package database

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/13SOAT-andromeda/tech-challenge-s1/internal/adapter/config"
	gormtrace "github.com/DataDog/dd-trace-go/contrib/gorm.io/gorm.v1/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type Database interface {
	AutoMigrate(models ...interface{}) error
	GetDB() *gorm.DB
}

type Postgres struct {
	db *gorm.DB
}

func (p *Postgres) AutoMigrate(models ...interface{}) error {
	return p.db.AutoMigrate(models...)
}

func (p *Postgres) GetDB() *gorm.DB {
	return p.db
}

func Init(ctx context.Context, config config.DataBaseConfig) (Database, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		config.Host,
		config.User,
		config.Password,
		config.DBName,
		config.Port,
		config.SSLMode,
		config.TimeZone)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})

	errCheck := gormtrace.WithErrorCheck(func(err error) bool {
		return !errors.Is(err, gorm.ErrRecordNotFound)
	})
	if err := db.Use(gormtrace.NewTracePlugin(errCheck)); err != nil {
		log.Fatal(err)
	}

	if err != nil {
		return nil, err
	}

	return &Postgres{db: db}, nil
}
