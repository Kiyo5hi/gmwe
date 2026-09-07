package storage_test

import (
	"gmwe/api/auth"
	"gmwe/api/db"
	"gmwe/api/hitokoto"
	"gmwe/api/storage"
	"gorm.io/gorm"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestRoundtripAndConstraints(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.db")
	engine, err := db.Open("file:" + filepath.ToSlash(path))
	if err != nil {
		t.Fatal(err)
	}
	pool, _ := engine.DB()
	defer pool.Close()
	stamp := time.Date(2026, 1, 2, 3, 4, 5, 123000000, time.UTC)
	input := storage.Snapshot{Users: []storage.UserRow{{ID: 2, CreatedAt: stamp, UpdatedAt: stamp, Username: "Alice", Name: "Fixture", Password: "hashed-fixture"}},
		Hitokotos: []storage.HitokotoRow{{ID: 7, CreatedAt: stamp, UpdatedAt: stamp, Content: "caf\u00e9", UserID: 2}}, NextUserID: 9, NextHitokotoID: 20}
	if err = storage.Import(engine, input); err != nil {
		t.Fatal(err)
	}
	if err = engine.AutoMigrate(&auth.User{}, &hitokoto.Hitokoto{}); err != nil {
		t.Fatal(err)
	}
	actual, err := storage.Read(engine)
	if err != nil || storage.Digest(actual) != storage.Digest(input) {
		t.Fatal("roundtrip differs", err)
	}
	if storage.Import(engine, input) == nil {
		t.Fatal("nonempty import accepted")
	}
	for _, text := range []string{"CAFE", "Cafe\u0301", "caf\u00e9 "} {
		if engine.Create(&hitokoto.Hitokoto{Content: text, UserID: 2}).Error == nil {
			t.Fatal("duplicate accepted")
		}
	}
	if engine.Create(&hitokoto.Hitokoto{Content: "orphan", UserID: 99}).Error == nil {
		t.Fatal("orphan accepted")
	}
	var relation hitokoto.Hitokoto
	if err = engine.Preload("User").First(&relation, 7).Error; err != nil || relation.User.ID != 2 {
		t.Fatal("association failed", err)
	}
	var wait sync.WaitGroup
	for i := 0; i < 8; i++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			var count int64
			if err := engine.Model(&auth.User{}).Count(&count).Error; err != nil {
				t.Error(err)
			}
		}()
	}
	wait.Wait()
	err = engine.Transaction(func(tx *gorm.DB) error {
		value := hitokoto.Hitokoto{Content: "transaction fixture", UserID: 2}
		if e := tx.Create(&value).Error; e != nil {
			return e
		}
		if value.ID != 20 {
			t.Error("sequence not preserved")
		}
		if e := tx.Model(&value).Update("Content", "updated fixture").Error; e != nil {
			return e
		}
		if e := tx.Delete(&value).Error; e != nil {
			return e
		}
		return gorm.ErrInvalidTransaction
	})
	if err != gorm.ErrInvalidTransaction {
		t.Fatal(err)
	}
	backup := filepath.Join(t.TempDir(), "backup.db")
	if err = storage.Backup(engine, backup); err != nil {
		t.Fatal(err)
	}
	restored, err := db.Open("file:" + filepath.ToSlash(backup) + "?mode=rw")
	if err != nil {
		t.Fatal(err)
	}
	restoredPool, _ := restored.DB()
	defer restoredPool.Close()
	got, err := storage.Read(restored)
	if err != nil || storage.Digest(got) != storage.Digest(input) {
		t.Fatal("backup differs", err)
	}
	if err = storage.Check(restored); err != nil {
		t.Fatal(err)
	}
}

func TestMissingDatabaseFailsClosed(t *testing.T) {
	engine, err := db.Open("file:" + filepath.ToSlash(filepath.Join(t.TempDir(), "missing.db")) + "?mode=rw")
	if err == nil {
		pool, _ := engine.DB()
		pool.Close()
		t.Fatal("missing database accepted")
	}
}
