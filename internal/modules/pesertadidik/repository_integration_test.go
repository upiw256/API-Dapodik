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
	err = pool.QueryRow(ctx, `WITH registration AS (SELECT DISTINCT ON (r.peserta_didik_id) r.peserta_didik_id,r.sekolah_id,r.jenis_keluar_id FROM public.registrasi_peserta_didik r WHERE r.soft_delete=0 ORDER BY r.peserta_didik_id,r.last_update DESC,r.registrasi_id DESC) SELECT pd.peserta_didik_id::text,reg.sekolah_id::text,rb.rombongan_belajar_id::text,btrim(rb.semester_id),rb.tingkat_pendidikan_id::bigint FROM public.peserta_didik pd JOIN registration reg ON reg.peserta_didik_id=pd.peserta_didik_id AND reg.jenis_keluar_id IS NULL JOIN public.anggota_rombel ar ON ar.peserta_didik_id=pd.peserta_didik_id AND ar.soft_delete=0 JOIN public.rombongan_belajar rb ON rb.rombongan_belajar_id=ar.rombongan_belajar_id AND rb.soft_delete=0 AND rb.sekolah_id=reg.sekolah_id AND rb.jenis_rombel IN (1,9) WHERE pd.soft_delete=0 ORDER BY rb.semester_id DESC,ar.last_update DESC LIMIT 1`).Scan(&studentID, &schoolID, &rombelID, &semesterID, &tingkatID)
	if err != nil {
		t.Fatalf("find registered student: %v", err)
	}
	repository := NewPostgresRepository(pool)
	items, total, err := repository.List(ctx, 1, 5, Filters{SekolahID: schoolID, RombonganBelajarID: rombelID, SemesterID: semesterID, TingkatPendidikanID: fmt.Sprint(tingkatID), Search: ""})
	if err != nil {
		t.Fatalf("list students: %v", err)
	}
	if total == 0 || len(items) == 0 {
		t.Fatal("expected registered students in selected school")
	}
	allItems, allTotal, err := repository.List(ctx, 1, 0, Filters{SekolahID: schoolID, RombonganBelajarID: rombelID, SemesterID: semesterID, TingkatPendidikanID: fmt.Sprint(tingkatID)})
	if err != nil {
		t.Fatalf("list all matching students without pagination: %v", err)
	}
	if int64(len(allItems)) != allTotal {
		t.Fatalf("unpaginated list returned %d rows but count query reports %d", len(allItems), allTotal)
	}
	for _, item := range allItems {
		if item.JenisRombel == nil || (*item.JenisRombel != 1 && *item.JenisRombel != 9) {
			t.Fatalf("student %s belongs to unexpected rombel type: %v", item.PesertaDidikID, item.JenisRombel)
		}
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
