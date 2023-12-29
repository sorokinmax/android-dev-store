package db

import (
	models "android-store/internal/models/apk"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SQLiteCreateDB create/magrate DB
func SQLiteCreateDB(ApkStruct models.Apk) error {
	db, err := gorm.Open(sqlite.Open("./data/apk.db"))
	if err != nil {
		panic("Failed to open the SQLite database.")
	}

	// Create the table from our struct.
	db.AutoMigrate(&ApkStruct)

	log.Println("Create/migrate registry-DB successfull")

	return nil
}

func SQLiteAddApk(apk *models.Apk) error {
	db, err := gorm.Open(sqlite.Open("./data/astore.db"))
	if err != nil {
		panic("Failed to open the SQLite database.")
	}

	db.Create(apk)

	return nil
}

func SQLiteDelApk(apk models.Apk) error {
	db, err := gorm.Open(sqlite.Open("./data/astore.db"))
	if err != nil {
		panic("Failed to open the SQLite database.")
	}

	db.Where("id = ?", apk.ID).Delete(&apk)

	return nil
}

func SQLiteSaveApk(apk models.Apk) error {
	db, err := gorm.Open(sqlite.Open("./data/astore.db"))
	if err != nil {
		panic("Failed to open the SQLite database.")
	}

	db.Where("id = ?", apk.ID).Save(&apk)

	return nil
}

func SQLiteGetApk(id string) (models.Apk, error) {
	db, err := gorm.Open(sqlite.Open("./data/astore.db"))
	if err != nil {
		panic("Failed to open the SQLite database.")
	}

	var apk models.Apk
	db.Where("id = ?", id).Find(&apk)

	return apk, nil
}

func SQLiteFindApk(sha256 string) (models.Apk, error) {
	db, err := gorm.Open(sqlite.Open("./data/astore.db"))
	if err != nil {
		panic("Failed to open the SQLite database.")
	}

	var apk models.Apk
	db.Where("sha256 = ?", sha256).Find(&apk)

	return apk, nil
}

func SQLiteGetApks() ([]models.Apk, error) {
	db, err := gorm.Open(sqlite.Open("./data/astore.db"))
	if err != nil {
		panic("Failed to open the SQLite database.")
	}

	var apks []models.Apk
	db.Order("created_at desc").Find(&apks)

	return apks, nil
}
