package sekolah

import (
	"context"
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

	repository := NewPostgresRepository(pool)
	schools, total, err := repository.ListSchools(ctx, 1, 3, "")
	if err != nil {
		t.Fatalf("list schools: %v", err)
	}
	if total == 0 || len(schools) == 0 {
		t.Fatal("expected at least one active school")
	}
	allSchools, allTotal, err := repository.ListSchools(ctx, 1, 0, "")
	if err != nil || int64(len(allSchools)) != allTotal {
		t.Fatalf("unpaginated school list returned %d rows, total=%d, err=%v", len(allSchools), allTotal, err)
	}
	if _, err := repository.GetSchool(ctx, schools[0].SekolahID); err != nil {
		t.Fatalf("get school: %v", err)
	}
	if npsn := os.Getenv("NPSN"); npsn != "" {
		filtered, filteredTotal, err := repository.ListSchools(ctx, 1, 3, npsn)
		if err != nil {
			t.Fatalf("list school by NPSN: %v", err)
		}
		if filteredTotal == 0 || len(filtered) == 0 || filtered[0].NPSN == nil || *filtered[0].NPSN != npsn {
			t.Fatalf("expected a school matching NPSN %s, got total=%d rows=%+v", npsn, filteredTotal, filtered)
		}
	}
	if _, err := repository.ListJurusan(ctx, schools[0].SekolahID); err != nil {
		t.Fatalf("list school majors: %v", err)
	}
	if _, err := repository.ListRombel(ctx, schools[0].SekolahID, ""); err != nil {
		t.Fatalf("list school rombel: %v", err)
	}
	if _, err := repository.ListRooms(ctx, schools[0].SekolahID); err != nil {
		t.Fatalf("list school rooms: %v", err)
	}

	var longitudinalSchoolID string
	err = pool.QueryRow(ctx, `
		SELECT sl.sekolah_id::text
		FROM public.sekolah_longitudinal AS sl
		JOIN public.sekolah AS s ON s.sekolah_id = sl.sekolah_id
		WHERE sl.soft_delete = 0
		  AND s.soft_delete = 0
		ORDER BY sl.semester_id DESC
		LIMIT 1`).Scan(&longitudinalSchoolID)
	if err != nil {
		t.Fatalf("find active longitudinal school: %v", err)
	}
	if _, err := repository.GetLongitudinal(ctx, longitudinalSchoolID, ""); err != nil {
		t.Fatalf("get school longitudinal data: %v", err)
	}
}
