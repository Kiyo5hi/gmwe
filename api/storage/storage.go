// Package storage implements portable, application-aware import and online backup.
package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	sqlite3 "github.com/mattn/go-sqlite3"
	"gmwe/api/auth"
	"gmwe/api/db"
	"gmwe/api/hitokoto"
	"gorm.io/gorm"
)

type UserRow struct {
	ID        uint
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
	Username  string
	Name      string
	Password  string
}
type HitokotoRow struct {
	ID                uint
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         gorm.DeletedAt
	Content           string
	UserID            int
	SubmittedByUserID *int `json:",omitempty"`
}
type Snapshot struct {
	Users          []UserRow     `json:"users"`
	Hitokotos      []HitokotoRow `json:"hitokotos"`
	NextUserID     int64         `json:"next_user_id"`
	NextHitokotoID int64         `json:"next_hitokoto_id"`
}

func Read(engine *gorm.DB) (Snapshot, error) {
	result := Snapshot{}
	err := engine.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Table("users").Order("id").Find(&result.Users).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Table("hitokotos").Order("id").Find(&result.Hitokotos).Error; err != nil {
			return err
		}
		if err := tx.Raw("SELECT COALESCE((SELECT seq FROM sqlite_sequence WHERE name='users'),0)+1").Scan(&result.NextUserID).Error; err != nil {
			return err
		}
		return tx.Raw("SELECT COALESCE((SELECT seq FROM sqlite_sequence WHERE name='hitokotos'),0)+1").Scan(&result.NextHitokotoID).Error
	})
	return result, err
}

func Digest(data Snapshot) string {
	encoded, _ := json.Marshal(data)
	value := sha256.Sum256(encoded)
	return hex.EncodeToString(value[:])
}

func Import(engine *gorm.DB, data Snapshot) error {
	if err := engine.AutoMigrate(&auth.User{}, &hitokoto.Hitokoto{}); err != nil {
		return err
	}
	return engine.Transaction(func(tx *gorm.DB) error {
		for _, table := range []string{"users", "hitokotos"} {
			var count int64
			if err := tx.Unscoped().Table(table).Count(&count).Error; err != nil {
				return err
			}
			if count != 0 {
				return fmt.Errorf("refusing nonempty target")
			}
		}
		if len(data.Users) != 0 {
			if err := tx.Table("users").Create(&data.Users).Error; err != nil {
				return err
			}
		}
		if len(data.Hitokotos) != 0 {
			if err := tx.Table("hitokotos").CreateInBatches(&data.Hitokotos, 100).Error; err != nil {
				return err
			}
		}
		for table, next := range map[string]int64{"users": data.NextUserID, "hitokotos": data.NextHitokotoID} {
			if err := tx.Exec("UPDATE sqlite_sequence SET seq=MAX(seq,?) WHERE name=?", next-1, table).Error; err != nil {
				return err
			}
		}
		restored, err := Read(tx)
		if err != nil {
			return err
		}
		if Digest(restored) != Digest(data) {
			return fmt.Errorf("import data or sequence mismatch")
		}
		return nil
	})
}

func Check(engine *gorm.DB) error {
	var integrity string
	if err := engine.Raw("PRAGMA integrity_check").Scan(&integrity).Error; err != nil {
		return err
	}
	if integrity != "ok" {
		return fmt.Errorf("integrity check failed")
	}
	var violations []map[string]interface{}
	if err := engine.Raw("PRAGMA foreign_key_check").Scan(&violations).Error; err != nil {
		return err
	}
	if len(violations) != 0 {
		return fmt.Errorf("foreign key check failed")
	}
	return nil
}

func Backup(engine *gorm.DB, destination string) error {
	absolute, err := filepath.Abs(destination)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(absolute, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	target, err := db.Open("file:" + filepath.ToSlash(absolute) + "?mode=rw")
	if err != nil {
		return err
	}
	targetPool, _ := target.DB()
	defer targetPool.Close()
	sourcePool, _ := engine.DB()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	sourceConn, err := sourcePool.Conn(ctx)
	if err != nil {
		return err
	}
	defer sourceConn.Close()
	targetConn, err := targetPool.Conn(ctx)
	if err != nil {
		return err
	}
	err = sourceConn.Raw(func(source interface{}) error {
		return targetConn.Raw(func(destination interface{}) error {
			backup, err := destination.(*sqlite3.SQLiteConn).Backup("main", source.(*sqlite3.SQLiteConn), "main")
			if err != nil {
				return err
			}
			defer backup.Close()
			done, err := backup.Step(-1)
			if err != nil {
				return err
			}
			if !done {
				return fmt.Errorf("backup incomplete")
			}
			return backup.Finish()
		})
	})
	targetConn.Close()
	if err != nil {
		return err
	}
	// Release the only source connection before subsequent verification reads.
	sourceConn.Close()
	if err = Check(target); err != nil {
		return err
	}
	return target.Exec("PRAGMA wal_checkpoint(TRUNCATE)").Error
}
