package pesertadidik

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
	var studentID, schoolID, rombelID, semesterID string
	var tingkatID int64
	err = pool.QueryRow(ctx, `SELECT pd.peserta_didik_id::text, reg.sekolah_id::text, rb.rombongan_belajar_id::text, btrim(rb.semester_id), rb.tingkat_pendidikan_id::bigint FROM public.peserta_didik AS pd JOIN public.registrasi_peserta_didik AS reg ON reg.peserta_didik_id = pd.peserta_didik_id AND reg.soft_delete = 0 AND reg.jenis_keluar_id IS NULL JOIN public.anggota_rombel AS ar ON ar.peserta_didik_id = pd.peserta_didik_id AND ar.soft_delete = 0 JOIN public.rombongan_belajar AS rb ON rb.rombongan_belajar_id = ar.rombongan_belajar_id AND rb.soft_delete = 0 WHERE pd.soft_delete = 0 ORDER BY pd.peserta_didik_id LIMIT 1`).Scan(&studentID, &schoolID, &rombelID, &semesterID, &tingkatID)
	if err != nil {
		t.Fatalf("find registered student: %v", err)
	}
	repository := NewPostgresRepository(pool)
	aktif := true
	items, total, err := repository.List(ctx, 1, 5, Filters{SekolahID: schoolID, RombonganBelajarID: rombelID, SemesterID: semesterID, TingkatPendidikanID: fmt.Sprint(tingkatID), Aktif: &aktif, Search: ""})
	if err != nil {
		t.Fatalf("list students: %v", err)
	}
	if total == 0 || len(items) == 0 {
		t.Fatal("expected registered students in selected school")
	}
	allItems, allTotal, err := repository.List(ctx, 1, 0, Filters{SekolahID: schoolID, RombonganBelajarID: rombelID, SemesterID: semesterID, TingkatPendidikanID: fmt.Sprint(tingkatID), Aktif: &aktif})
	if err != nil {
		t.Fatalf("list all matching students without pagination: %v", err)
	}
	if int64(len(allItems)) != allTotal {
		t.Fatalf("unpaginated list returned %d rows but count query reports %d", len(allItems), allTotal)
	}
	if _, err := repository.GetDetail(ctx, studentID); err != nil {
		t.Fatalf("get student detail: %v", err)
	}
	if _, err := repository.ListRombel(ctx, studentID); err != nil {
		t.Fatalf("list student rombel history: %v", err)
	}
	if _, err := repository.GetPrestasiBeasiswa(ctx, studentID); err != nil {
		t.Fatalf("get student achievements and scholarships: %v", err)
	}
}
