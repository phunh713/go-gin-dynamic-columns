package config

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const ContextKeyDB = "db"

func NewDB(configEnv *ConfigEnv) *gorm.DB {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		configEnv.DbHost,
		configEnv.DbUsername,
		configEnv.DbPassword,
		configEnv.DbDatabase,
		configEnv.DbPort)
	fmt.Println(dsn)

	// Configure logger to print all SQL queries
	// newLogger := logger.New(
	// 	log.New(os.Stdout, "\r\n", log.LstdFlags),
	// 	logger.Config{
	// 		SlowThreshold:             time.Second,
	// 		LogLevel:                  logger.Info,
	// 		IgnoreRecordNotFoundError: true,
	// 		Colorful:                  true,
	// 	},
	// )
	// newLogger = nil
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic(err)
	}
	return db
}
