package database

import (
	"fmt"

	"github.com/s8508235/tui-dictionary/pkg/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewSqlLiteConnection(dbName string, logger *log.Logger) (*gorm.DB, error) {
	db := fmt.Sprintf("%s.db", dbName)
	return gorm.Open(sqlite.Open(db), &gorm.Config{
		Logger: logger,
	})
}
