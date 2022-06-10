package database

import (
	"log"
	"os"

	"github.com/its-me-debk007/auth-backend/models"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err.Error())
	}
	//dbPassword := os.Getenv("DB_PASSWORD")
	//dbName := os.Getenv("DB_NAME")
	//dbUsername := os.Getenv("DB_USERNAME")
	//
	//dbUrl := fmt.Sprintf("postgresql://%s:%s@localhost/%s?sslmode=disable", dbUsername, dbPassword, dbName)

	dbUrl := os.Getenv("DATABASE_URL")

	db, err := gorm.Open(postgres.Open(dbUrl), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	DB = db
	if err := db.AutoMigrate(new(models.User), new(models.Otps)); err != nil {
		log.Fatal(err.Error())
	}
}
