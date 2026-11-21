package ptk

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"dapodik-readonly-api/internal/config"
	"dapodik-readonly-api/internal/database"
	"github.com/joho/godotenv"
)

func TestPostgresRepositoryIntegration(t *testing.T) {
	if os.Getenv("DAPODIK_TEST_DB") != "1" {
		t.Skip("set DAPODIK_TEST_DB=1 to run read-only database integration checks")
	}
	if err := godotenv.Load("../../../.env"); err != nil {
		t.Fatalf("load project .env: %v", err)
	}
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := database.Open(ctx, cfg.DBDSN, cfg.DBConnectTimeout, cfg.DBMinConns, cfg.DBMaxConns)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer pool.Close()

	var ptkID, sekolahID string
	var tahunAjaranID, jenisPTKID int64
	err = pool.QueryRow(ctx, `
		SELECT ptk.ptk_id::text,
	       terdaftar.sekolah_id::text,
	       terdaftar.tahun_ajaran_id::bigint,
	       terdaftar.jenis_ptk_id::bigint
	FROM public.ptk AS ptk
	JOIN public.ptk_terdaftar AS terdaftar ON terdaftar.ptk_id = ptk.ptk_id
	WHERE ptk.soft_delete = 0
	  AND terdaftar.soft_delete = 0
	ORDER BY ptk.ptk_id, terdaftar.tahun_ajaran_id DESC
	LIMIT 1`).Scan(&ptkID, &sekolahID, &tahunAjaranID, &jenisPTKID)
	if err != nil {
		t.Fatalf("find active ptk: %v", err)
	}

	repository := NewPostgresRepository(pool)
	items, total, err := repository.List(ctx, 1, 5, Filters{
		SekolahID:     sekolahID,
		TahunAjaranID: fmt.Sprint(tahunAjaranID),
		JenisPTKID:    fmt.Sprint(jenisPTKID),
	})
	if err != nil {
		t.Fatalf("list ptk with filters: %v", err)
	}
	if total == 0 || len(items) == 0 {
		t.Fatal("expected at least one PTK matching the selected school and period")
	}
	allItems, allTotal, err := repository.List(ctx, 1, 0, Filters{SekolahID: sekolahID, TahunAjaranID: fmt.Sprint(tahunAjaranID), JenisPTKID: fmt.Sprint(jenisPTKID)})
	if err != nil || int64(len(allItems)) != allTotal {
		t.Fatalf("unpaginated PTK list returned %d rows, total=%d, err=%v", len(allItems), allTotal, err)
	}
	if _, err := repository.GetDetail(ctx, ptkID); err != nil {
		t.Fatalf("get ptk detail: %v", err)
	}
	if _, err := repository.ListPembelajaran(ctx, ptkID, ""); err != nil {
		t.Fatalf("list ptk pembelajaran: %v", err)
	}
	if _, err := repository.ListTugasTambahan(ctx, ptkID); err != nil {
		t.Fatalf("list ptk tugas tambahan: %v", err)
	}
	if _, err := repository.GetRiwayat(ctx, ptkID); err != nil {
		t.Fatalf("get ptk riwayat: %v", err)
	}
}
