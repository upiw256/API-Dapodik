package referensi

import (
	"context"
	"fmt"

	"dapodik-readonly-api/internal/period"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store interface {
	ListSemesters(context.Context) ([]Semester, error)
	ListTahunAjaran(context.Context) ([]TahunAjaran, error)
	GetWilayahHierarchy(context.Context, string) (WilayahHierarchy, error)
}

type PostgresRepository struct {
	pool   *pgxpool.Pool
	period period.Selection
}

func NewPostgresRepository(pool *pgxpool.Pool, selected ...period.Selection) *PostgresRepository {
	r := &PostgresRepository{pool: pool}
	if len(selected) > 0 {
		r.period = selected[0]
	}
	return r
}

func (r *PostgresRepository) ListSemesters(ctx context.Context) ([]Semester, error) {
	const query = `
		SELECT btrim(semester_id),
		       tahun_ajaran_id::bigint,
		       nama,
		       semester::integer,
		       periode_aktif::integer,
		       to_char(tanggal_mulai, 'YYYY-MM-DD'),
		       to_char(tanggal_selesai, 'YYYY-MM-DD')
		FROM ref.semester
		WHERE expired_date IS NULL AND ($1::text = '' OR btrim(semester_id) = $1::text)
		ORDER BY tahun_ajaran_id DESC, semester DESC`

	rows, err := r.pool.Query(ctx, query, r.period.SemesterID)
	if err != nil {
		return nil, fmt.Errorf("query semesters: %w", err)
	}
	defer rows.Close()

	items := make([]Semester, 0)
	for rows.Next() {
		var item Semester
		if err := rows.Scan(
			&item.SemesterID,
			&item.TahunAjaranID,
			&item.Nama,
			&item.Semester,
			&item.PeriodeAktif,
			&item.TanggalMulai,
			&item.TanggalSelesai,
		); err != nil {
			return nil, fmt.Errorf("scan semester: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate semesters: %w", err)
	}
	return items, nil
}

func (r *PostgresRepository) ListTahunAjaran(ctx context.Context) ([]TahunAjaran, error) {
	const query = `
		SELECT tahun_ajaran_id::bigint,
		       nama,
		       periode_aktif::integer,
		       to_char(tanggal_mulai, 'YYYY-MM-DD'),
		       to_char(tanggal_selesai, 'YYYY-MM-DD')
		FROM ref.tahun_ajaran
		WHERE expired_date IS NULL AND ($1::bigint = 0 OR tahun_ajaran_id = $1::bigint)
		ORDER BY tahun_ajaran_id DESC`

	rows, err := r.pool.Query(ctx, query, r.period.TahunAjaranID)
	if err != nil {
		return nil, fmt.Errorf("query tahun ajaran: %w", err)
	}
	defer rows.Close()

	items := make([]TahunAjaran, 0)
	for rows.Next() {
		var item TahunAjaran
		if err := rows.Scan(
			&item.TahunAjaranID,
			&item.Nama,
			&item.PeriodeAktif,
			&item.TanggalMulai,
			&item.TanggalSelesai,
		); err != nil {
			return nil, fmt.Errorf("scan tahun ajaran: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tahun ajaran: %w", err)
	}
	return items, nil
}

func (r *PostgresRepository) GetWilayahHierarchy(ctx context.Context, kode string) (WilayahHierarchy, error) {
	const query = `
		WITH RECURSIVE wilayah_hierarki AS (
			SELECT kode_wilayah, nama, id_level_wilayah, mst_kode_wilayah, 0 AS depth
			FROM ref.mst_wilayah
			WHERE kode_wilayah = $1
			  AND expired_date IS NULL

			UNION ALL

			SELECT parent.kode_wilayah,
			       parent.nama,
			       parent.id_level_wilayah,
			       parent.mst_kode_wilayah,
			       child.depth + 1
			FROM ref.mst_wilayah AS parent
			JOIN wilayah_hierarki AS child
			  ON parent.kode_wilayah = child.mst_kode_wilayah
			WHERE parent.expired_date IS NULL
			  AND child.depth < 4
		)
		SELECT btrim(kode_wilayah), nama, id_level_wilayah
		FROM wilayah_hierarki
		WHERE id_level_wilayah BETWEEN 1 AND 4
		ORDER BY id_level_wilayah`

	rows, err := r.pool.Query(ctx, query, kode)
	if err != nil {
		return WilayahHierarchy{}, fmt.Errorf("query wilayah hierarchy: %w", err)
	}
	defer rows.Close()

	var result WilayahHierarchy
	found := false
	for rows.Next() {
		found = true
		var kodeWilayah, nama string
		var level int16
		if err := rows.Scan(&kodeWilayah, &nama, &level); err != nil {
			return WilayahHierarchy{}, fmt.Errorf("scan wilayah hierarchy: %w", err)
		}
		switch level {
		case 1:
			result.Provinsi = &Region{KodeWilayah: kodeWilayah, Nama: nama}
		case 2:
			result.KabupatenKota = &Region{KodeWilayah: kodeWilayah, Nama: nama}
		case 3:
			result.Kecamatan = &Region{KodeWilayah: kodeWilayah, Nama: nama}
		case 4:
			result.DesaKelurahan = &Region{KodeWilayah: kodeWilayah, Nama: nama}
		}
	}
	if err := rows.Err(); err != nil {
		return WilayahHierarchy{}, fmt.Errorf("iterate wilayah hierarchy: %w", err)
	}
	if !found {
		return WilayahHierarchy{}, ErrNotFound
	}
	return result, nil
}
