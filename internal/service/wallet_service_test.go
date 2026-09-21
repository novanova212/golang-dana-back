package service

// Berbeda dari bill_service_test.go (yang pakai fake repository), test
// ini sengaja pakai DATABASE BENERAN (SQLite in-memory) karena logic
// Transfer() sangat bergantung pada fitur database transaction & row
// locking dari GORM (`db.Transaction(...)`) - fitur ini tidak bisa
// disimulasikan dengan fake repository biasa, harus diuji lewat
// database sungguhan (SQLite in-memory dipilih karena cepat dan tidak
// perlu instalasi PostgreSQL terpisah hanya untuk testing).
//
// Jalankan dengan: go test ./internal/service/... -v
// Perlu: go get github.com/glebarez/sqlite
//
// Catatan: kita pakai github.com/glebarez/sqlite (bukan gorm.io/driver/sqlite
// bawaan) karena driver bawaan butuh CGO + compiler C terinstall di sistem,
// yang seringkali TIDAK ada di Windows secara default. Driver glebarez ini
// pure Go (tanpa CGO), jadi langsung jalan di komputer mana pun tanpa
// perlu install compiler C tambahan.

import (
	"fmt"
	"testing"

	"dana-clone/internal/model"
	"dana-clone/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// setupTestDB membuat database SQLite in-memory yang SEGAR & TERISOLASI
// untuk setiap test. Nama database dibuat unik berdasarkan nama test
// (t.Name()) - ini penting! Tanpa nama unik, semua test akan berbagi
// database in-memory yang sama (karena mode "cache=shared" mempertahankan
// koneksi tetap hidup antar test), menyebabkan data dari 1 test "bocor"
// ke test lain (contoh: insert user dengan email yang sama gagal karena
// dianggap duplikat, padahal itu 2 test yang seharusnya independen).
func setupTestDB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&model.User{}, &model.Transaction{})
	require.NoError(t, err)

	return db
}

func TestTopUp_Sukses(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	txRepo := repository.NewTransactionRepository(db)
	svc := NewWalletService(userRepo, txRepo, db)

	db.Create(&model.User{Name: "Budi", Email: "budi@test.com", Balance: 0})

	newBalance, err := svc.TopUp(1, 50000)

	assert.NoError(t, err)
	assert.Equal(t, int64(50000), newBalance)
}

func TestTopUp_JumlahInvalid(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	txRepo := repository.NewTransactionRepository(db)
	svc := NewWalletService(userRepo, txRepo, db)

	db.Create(&model.User{Name: "Budi", Email: "budi@test.com", Balance: 0})

	_, err := svc.TopUp(1, -1000)
	assert.Error(t, err)

	_, err = svc.TopUp(1, 0)
	assert.Error(t, err)
}

func TestTransfer_Sukses(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	txRepo := repository.NewTransactionRepository(db)
	svc := NewWalletService(userRepo, txRepo, db)

	db.Create(&model.User{Name: "Budi", Email: "budi@test.com", Balance: 100000})
	db.Create(&model.User{Name: "Ani", Email: "ani@test.com", Balance: 0})

	err := svc.Transfer(1, 2, 40000)
	assert.NoError(t, err)

	sender, _ := userRepo.FindByID(1)
	receiver, _ := userRepo.FindByID(2)
	assert.Equal(t, int64(60000), sender.Balance)
	assert.Equal(t, int64(40000), receiver.Balance)
}

func TestTransfer_SaldoTidakCukup(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	txRepo := repository.NewTransactionRepository(db)
	svc := NewWalletService(userRepo, txRepo, db)

	db.Create(&model.User{Name: "Budi", Email: "budi@test.com", Balance: 10000})
	db.Create(&model.User{Name: "Ani", Email: "ani@test.com", Balance: 0})

	err := svc.Transfer(1, 2, 50000)
	assert.Error(t, err)

	// PENTING: pastikan saldo TIDAK BERUBAH SAMA SEKALI meski transfer
	// gagal di tengah jalan - ini membuktikan database transaction
	// (rollback) bekerja dengan benar.
	sender, _ := userRepo.FindByID(1)
	receiver, _ := userRepo.FindByID(2)
	assert.Equal(t, int64(10000), sender.Balance)
	assert.Equal(t, int64(0), receiver.Balance)
}

func TestTransfer_KeDiriSendiri(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	txRepo := repository.NewTransactionRepository(db)
	svc := NewWalletService(userRepo, txRepo, db)

	db.Create(&model.User{Name: "Budi", Email: "budi@test.com", Balance: 100000})

	err := svc.Transfer(1, 1, 10000)
	assert.Error(t, err)
}

func TestTransfer_PenerimaTidakDitemukan(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	txRepo := repository.NewTransactionRepository(db)
	svc := NewWalletService(userRepo, txRepo, db)

	db.Create(&model.User{Name: "Budi", Email: "budi@test.com", Balance: 100000})

	err := svc.Transfer(1, 999, 10000)
	assert.Error(t, err)

	// Saldo pengirim tidak boleh berkurang meski penerimanya tidak valid.
	sender, _ := userRepo.FindByID(1)
	assert.Equal(t, int64(100000), sender.Balance)
}

func TestGetHistory_Pagination(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	txRepo := repository.NewTransactionRepository(db)
	svc := NewWalletService(userRepo, txRepo, db)

	db.Create(&model.User{Name: "Budi", Email: "budi@test.com", Balance: 0})

	// Bikin 5 transaksi top up.
	for i := 0; i < 5; i++ {
		svc.TopUp(1, 10000)
	}

	page1, total, err := svc.GetHistory(1, "", "", 1, 2)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, page1, 2, "halaman 1 dengan limit 2 harus berisi 2 data")

	page3, _, err := svc.GetHistory(1, "", "", 3, 2)
	assert.NoError(t, err)
	assert.Len(t, page3, 1, "halaman terakhir cuma sisa 1 data dari total 5")
}