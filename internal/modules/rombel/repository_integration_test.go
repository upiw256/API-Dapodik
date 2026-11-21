package rombel

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
	var id, schoolID, semesterID string
	err = pool.QueryRow(ctx, `SELECT rb.rombongan_belajar_id::text, rb.sekolah_id::text, btrim(rb.semester_id) FROM public.rombongan_belajar AS rb WHERE rb.soft_delete=0 ORDER BY rb.rombongan_belajar_id LIMIT 1`).Scan(&id, &schoolID, &semesterID)
	if err != nil {
		t.Fatalf("find rombel: %v", err)
	}
	repository := NewPostgresRepository(pool)
	items, total, err := repository.List(ctx, 1, 5, Filters{SekolahID: schoolID, SemesterID: semesterID})
	if err != nil {
		t.Fatalf("list rombel: %v", err)
	}
	if total == 0 || len(items) == 0 {
		t.Fatal("expected a rombel in selected school and semester")
	}
	allItems, allTotal, err := repository.List(ctx, 1, 0, Filters{SekolahID: schoolID, SemesterID: semesterID})
	if err != nil || int64(len(allItems)) != allTotal {
		t.Fatalf("unpaginated rombel list returned %d rows, total=%d, err=%v", len(allItems), allTotal, err)
	}
	rows, err := pool.Query(ctx, `SELECT DISTINCT jenis_rombel::bigint FROM public.rombongan_belajar WHERE soft_delete = 0 ORDER BY jenis_rombel::bigint`)
	if err != nil {
		t.Fatalf("list available rombel types: %v", err)
	}
	var jenisList []int64
	for rows.Next() {
		var jenis int64
		if err := rows.Scan(&jenis); err != nil {
			rows.Close()
			t.Fatalf("scan rombel type: %v", err)
		}
		jenisList = append(jenisList, jenis)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		t.Fatalf("iterate rombel types: %v", err)
	}
	rows.Close()
	for _, jenis := range jenisList {
		if _, _, err := repository.List(ctx, 1, 5, Filters{JenisRombel: fmt.Sprint(jenis)}); err != nil {
			t.Fatalf("list rombel type %d: %v", jenis, err)
		}
	}
	if _, err := repository.GetDetail(ctx, id); err != nil {
		t.Fatalf("get rombel detail: %v", err)
	}
	if _, err := repository.ListAnggota(ctx, id); err != nil {
		t.Fatalf("list rombel members: %v", err)
	}
	if _, err := repository.ListPembelajaran(ctx, id); err != nil {
		t.Fatalf("list rombel learning: %v", err)
	}
	if _, err := repository.GetStatistics(ctx, id); err != nil {
		t.Fatalf("get rombel statistics: %v", err)
	}
}
