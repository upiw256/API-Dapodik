package rombel

import (
	"context"
	"errors"
	"fmt"

	"dapodik-readonly-api/internal/period"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("rombel tidak ditemukan")

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

const activeStudentCounts = `
WITH current_registration AS (
	SELECT DISTINCT ON (r.peserta_didik_id)
	       r.peserta_didik_id, r.jenis_keluar_id
	FROM public.registrasi_peserta_didik AS r
	WHERE r.soft_delete = 0
	ORDER BY r.peserta_didik_id, r.last_update DESC, r.registrasi_id DESC
), student_counts AS (
	SELECT ar.rombongan_belajar_id, count(DISTINCT ar.peserta_didik_id)::bigint AS jumlah
	FROM public.anggota_rombel AS ar
	JOIN public.peserta_didik AS pd ON pd.peserta_didik_id = ar.peserta_didik_id AND pd.soft_delete = 0
	JOIN current_registration AS reg ON reg.peserta_didik_id = pd.peserta_didik_id AND reg.jenis_keluar_id IS NULL
	WHERE ar.soft_delete = 0
	GROUP BY ar.rombongan_belajar_id
)`

const rombelFilter = `
WHERE rb.soft_delete = 0
	AND rb.jenis_rombel NOT IN (18, 51)
	AND ($1::text = '' OR rb.sekolah_id::text = $1::text)
	AND ($2::text = '' OR btrim(rb.semester_id) = $2::text)
	AND ($3::text = '' OR rb.tingkat_pendidikan_id::text = $3::text)
	AND ($4::text = '' OR rb.jenis_rombel::text = $4::text)
	AND ($5::text = '' OR rb.kurikulum_id::text = $5::text)
	AND ($6::text = '' OR rb.jurusan_sp_id::text = $6::text)
	AND ($7::text = '' OR rb.ptk_id::text = $7::text)
	AND ($8::text = '' OR rb.nama ILIKE '%' || $8::text || '%')`

func filterArgs(f Filters) []any {
	return []any{f.SekolahID, f.SemesterID, f.TingkatPendidikanID, f.JenisRombel, f.KurikulumID, f.JurusanSPID, f.PTKID, f.Search}
}

func (r *PostgresRepository) List(ctx context.Context, page, limit int, f Filters) ([]ListItem, int64, error) {
	if r.period.SemesterID != "" {
		f.SemesterID = r.period.SemesterID
	}
	args := filterArgs(f)
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM public.rombongan_belajar AS rb `+rombelFilter, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count rombel: %w", err)
	}
	query := activeStudentCounts + `
	SELECT rb.rombongan_belajar_id::text, rb.nama, btrim(rb.semester_id), semester.nama,
	       rb.sekolah_id::text, sekolah.nama, rb.tingkat_pendidikan_id::bigint, tingkat.nama,
	       rb.jenis_rombel::bigint, jenis.nm_jenis_rombel, rb.kurikulum_id::bigint, kurikulum.nama_kurikulum,
	       rb.jurusan_sp_id::text, jurusan_sp.nama_jurusan_sp, jurusan.nama_jurusan,
	       rb.ptk_id::text, ptk.nama, rb.id_ruang::text, ruang.nm_ruang, COALESCE(sc.jumlah, 0)::bigint
	FROM public.rombongan_belajar AS rb
	LEFT JOIN ref.semester AS semester ON btrim(semester.semester_id) = btrim(rb.semester_id) AND semester.expired_date IS NULL
	LEFT JOIN public.sekolah AS sekolah ON sekolah.sekolah_id = rb.sekolah_id AND sekolah.soft_delete = 0
	LEFT JOIN ref.tingkat_pendidikan AS tingkat ON tingkat.tingkat_pendidikan_id = rb.tingkat_pendidikan_id AND tingkat.expired_date IS NULL
	LEFT JOIN ref.jenis_rombel AS jenis ON jenis.jenis_rombel = rb.jenis_rombel AND jenis.expired_date IS NULL
	LEFT JOIN ref.kurikulum AS kurikulum ON kurikulum.kurikulum_id = rb.kurikulum_id AND kurikulum.expired_date IS NULL
	LEFT JOIN public.jurusan_sp AS jurusan_sp ON jurusan_sp.jurusan_sp_id = rb.jurusan_sp_id AND jurusan_sp.soft_delete = 0
	LEFT JOIN ref.jurusan AS jurusan ON jurusan.jurusan_id = jurusan_sp.jurusan_id AND jurusan.expired_date IS NULL
	LEFT JOIN public.ptk AS ptk ON ptk.ptk_id = rb.ptk_id AND ptk.soft_delete = 0
	LEFT JOIN public.ruang AS ruang ON ruang.id_ruang = rb.id_ruang AND ruang.soft_delete = 0
	LEFT JOIN student_counts AS sc ON sc.rombongan_belajar_id = rb.rombongan_belajar_id
	` + rombelFilter + `
	ORDER BY rb.nama, rb.rombongan_belajar_id`
	if limit > 0 {
		query += ` LIMIT $9 OFFSET $10`
		args = append(args, limit, (page-1)*limit)
	}
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query rombel: %w", err)
	}
	defer rows.Close()
	items := make([]ListItem, 0)
	for rows.Next() {
		var item ListItem
		if err := rows.Scan(&item.ID, &item.Nama, &item.SemesterID, &item.Semester, &item.SekolahID, &item.Sekolah, &item.TingkatPendidikanID, &item.TingkatPendidikan, &item.JenisRombel, &item.JenisRombelNama, &item.KurikulumID, &item.Kurikulum, &item.JurusanSPID, &item.JurusanSP, &item.Jurusan, &item.PTKID, &item.WaliKelas, &item.RuangID, &item.Ruang, &item.JumlahSiswa); err != nil {
			return nil, 0, fmt.Errorf("scan rombel: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate rombel: %w", err)
	}
	return items, total, nil
}

func (r *PostgresRepository) GetDetail(ctx context.Context, id string) (Detail, error) {
	const query = `
	SELECT rb.rombongan_belajar_id::text, rb.nama, btrim(rb.semester_id), semester.nama,
	       rb.sekolah_id::text, sekolah.nama, rb.tingkat_pendidikan_id::bigint, tingkat.nama,
	       rb.kurikulum_id::bigint, kurikulum.nama_kurikulum, rb.jenis_rombel::bigint, jenis.nm_jenis_rombel,
	       rb.moving_class::bigint, rb.sks::double precision, to_char(rb.tanggal_mulai, 'YYYY-MM-DD'), to_char(rb.tanggal_selesai, 'YYYY-MM-DD'),
	       rb.kebutuhan_khusus_id, kk.kebutuhan_khusus,
	       rb.ptk_id::text, ptk.nama, rb.id_ruang::text, ruang.nm_ruang, prasarana.nama,
	       bangunan.id_bangunan::text, bangunan.nama, rb.jurusan_sp_id::text, jurusan_sp.nama_jurusan_sp, jurusan.nama_jurusan,
	       COALESCE(sc.jumlah, 0)::bigint, validasi.status_validasi::bigint, COALESCE(validasi.jumlah, 0)::bigint,
	       FALSE
	FROM public.rombongan_belajar AS rb
	LEFT JOIN ref.semester AS semester ON btrim(semester.semester_id) = btrim(rb.semester_id) AND semester.expired_date IS NULL
	LEFT JOIN public.sekolah AS sekolah ON sekolah.sekolah_id = rb.sekolah_id AND sekolah.soft_delete = 0
	LEFT JOIN ref.tingkat_pendidikan AS tingkat ON tingkat.tingkat_pendidikan_id = rb.tingkat_pendidikan_id AND tingkat.expired_date IS NULL
	LEFT JOIN ref.kurikulum AS kurikulum ON kurikulum.kurikulum_id = rb.kurikulum_id AND kurikulum.expired_date IS NULL
	LEFT JOIN ref.jenis_rombel AS jenis ON jenis.jenis_rombel = rb.jenis_rombel AND jenis.expired_date IS NULL
	LEFT JOIN ref.kebutuhan_khusus AS kk ON kk.kebutuhan_khusus_id = rb.kebutuhan_khusus_id AND kk.expired_date IS NULL
	LEFT JOIN public.ptk AS ptk ON ptk.ptk_id = rb.ptk_id AND ptk.soft_delete = 0
	LEFT JOIN public.ruang AS ruang ON ruang.id_ruang = rb.id_ruang AND ruang.soft_delete = 0
	LEFT JOIN ref.jenis_prasarana AS prasarana ON prasarana.jenis_prasarana_id = ruang.jenis_prasarana_id AND prasarana.expired_date IS NULL
	LEFT JOIN public.bangunan AS bangunan ON bangunan.id_bangunan = ruang.id_bangunan AND bangunan.soft_delete = 0
	LEFT JOIN public.jurusan_sp AS jurusan_sp ON jurusan_sp.jurusan_sp_id = rb.jurusan_sp_id AND jurusan_sp.soft_delete = 0
	LEFT JOIN ref.jurusan AS jurusan ON jurusan.jurusan_id = jurusan_sp.jurusan_id AND jurusan.expired_date IS NULL
	LEFT JOIN student_counts AS sc ON sc.rombongan_belajar_id = rb.rombongan_belajar_id
	LEFT JOIN (
		SELECT v.rombongan_belajar_id, max(v.status_validasi)::bigint AS status_validasi,
		       count(*)::bigint AS jumlah
		FROM public.vld_rombel AS v
		WHERE v.soft_delete = 0
		GROUP BY v.rombongan_belajar_id
	) AS validasi ON validasi.rombongan_belajar_id = rb.rombongan_belajar_id
	WHERE rb.rombongan_belajar_id = $1::uuid AND rb.soft_delete = 0
	  AND ($2::text = '' OR btrim(rb.semester_id) = $2::text)`
	var item Detail
	err := r.pool.QueryRow(ctx, activeStudentCounts+query, id, r.period.SemesterID).Scan(&item.ID, &item.Nama, &item.SemesterID, &item.Semester, &item.SekolahID, &item.Sekolah, &item.TingkatPendidikanID, &item.TingkatPendidikan, &item.KurikulumID, &item.Kurikulum, &item.JenisRombel, &item.JenisRombelNama, &item.MovingClass, &item.SKS, &item.TanggalMulai, &item.TanggalSelesai, &item.KebutuhanKhususID, &item.KebutuhanKhusus, &item.PTKID, &item.WaliKelas, &item.RuangID, &item.Ruang, &item.JenisPrasarana, &item.BangunanID, &item.Bangunan, &item.JurusanSPID, &item.JurusanSP, &item.Jurusan, &item.JumlahSiswa, &item.StatusValidasi, &item.JumlahValidasi, &item.EkstrakurikulerTersedia)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Detail{}, ErrNotFound
		}
		return Detail{}, fmt.Errorf("query detail rombel: %w", err)
	}
	return item, nil
}

func (r *PostgresRepository) ListAnggota(ctx context.Context, id string) (AnggotaResponse, error) {
	result := AnggotaResponse{RombelID: id, Anggota: make([]Anggota, 0)}
	if err := r.pool.QueryRow(ctx, `SELECT rb.nama, ptk.nama FROM public.rombongan_belajar AS rb LEFT JOIN public.ptk AS ptk ON ptk.ptk_id=rb.ptk_id AND ptk.soft_delete=0 WHERE rb.rombongan_belajar_id=$1::uuid AND rb.soft_delete=0 AND ($2::text = '' OR btrim(rb.semester_id) = $2::text)`, id, r.period.SemesterID).Scan(&result.NamaKelas, &result.WaliKelas); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AnggotaResponse{}, ErrNotFound
		}
		return AnggotaResponse{}, fmt.Errorf("query metadata rombel anggota: %w", err)
	}
	const query = `
	SELECT ar.anggota_rombel_id::text, pd.peserta_didik_id::text, pd.nama, btrim(pd.jenis_kelamin), NULLIF(btrim(pd.nisn), ''),
	       reg.nipd, to_char(reg.tanggal_masuk_sekolah, 'YYYY-MM-DD'), ar.jenis_pendaftaran_id::bigint, pda.nama,
	       awal.jenis_pendaftaran_id::bigint, awal_ref.nama, pd.agama_id::integer, agama.nama,
	       reg.jenis_keluar_id, keluar.ket_keluar
	FROM public.anggota_rombel AS ar
	JOIN public.rombongan_belajar AS rb ON rb.rombongan_belajar_id = ar.rombongan_belajar_id AND rb.soft_delete = 0
	JOIN public.peserta_didik AS pd ON pd.peserta_didik_id = ar.peserta_didik_id AND pd.soft_delete = 0
	LEFT JOIN (
		SELECT DISTINCT ON (r.peserta_didik_id, r.sekolah_id)
		       r.peserta_didik_id, r.sekolah_id, r.nipd, r.tanggal_masuk_sekolah,
		       r.jenis_keluar_id
		FROM public.registrasi_peserta_didik AS r
		WHERE r.soft_delete = 0
		ORDER BY r.peserta_didik_id, r.sekolah_id, r.last_update DESC, r.registrasi_id DESC
	) AS reg ON reg.peserta_didik_id = pd.peserta_didik_id AND reg.sekolah_id = rb.sekolah_id
	LEFT JOIN (
		SELECT DISTINCT ON (r.peserta_didik_id)
		       r.peserta_didik_id, r.jenis_pendaftaran_id
		FROM public.registrasi_peserta_didik AS r
		WHERE r.soft_delete = 0
		ORDER BY r.peserta_didik_id, r.tanggal_masuk_sekolah ASC NULLS LAST, r.create_date ASC
	) AS awal ON awal.peserta_didik_id = pd.peserta_didik_id

	LEFT JOIN ref.jenis_pendaftaran AS pda ON pda.jenis_pendaftaran_id = ar.jenis_pendaftaran_id AND pda.expired_date IS NULL
	LEFT JOIN ref.jenis_pendaftaran AS awal_ref ON awal_ref.jenis_pendaftaran_id = awal.jenis_pendaftaran_id AND awal_ref.expired_date IS NULL
	LEFT JOIN ref.agama AS agama ON agama.agama_id = pd.agama_id AND agama.expired_date IS NULL
	LEFT JOIN ref.jenis_keluar AS keluar ON keluar.jenis_keluar_id = reg.jenis_keluar_id AND keluar.expired_date IS NULL
	WHERE ar.rombongan_belajar_id = $1::uuid AND ar.soft_delete = 0
	ORDER BY pd.nama ASC, pd.peserta_didik_id`
	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		return AnggotaResponse{}, fmt.Errorf("query anggota rombel: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item Anggota
		if err := rows.Scan(&item.AnggotaRombelID, &item.PesertaDidikID, &item.Nama, &item.JenisKelamin, &item.NISN, &item.NIPD, &item.TanggalMasuk, &item.JenisPendaftaranRombelID, &item.JenisPendaftaranRombel, &item.JenisPendaftaranAwalID, &item.JenisPendaftaranAwal, &item.AgamaID, &item.Agama, &item.JenisKeluarID, &item.JenisKeluar); err != nil {
			return AnggotaResponse{}, fmt.Errorf("scan anggota rombel: %w", err)
		}
		result.Anggota = append(result.Anggota, item)
	}
	if err := rows.Err(); err != nil {
		return AnggotaResponse{}, fmt.Errorf("iterate anggota rombel: %w", err)
	}
	return result, nil
}

func (r *PostgresRepository) ListPembelajaran(ctx context.Context, id string) ([]Pembelajaran, error) {
	const query = `
	SELECT pb.pembelajaran_id::text, btrim(pb.semester_id), pb.mata_pelajaran_id, NULL::text,
	       COALESCE(mp.nama, pb.nama_mata_pelajaran), ptk.ptk_id::text, ptk.nama, terdaftar.ptk_terdaftar_id::text,
	       pb.sk_mengajar, to_char(pb.tanggal_sk_mengajar, 'YYYY-MM-DD'), pb.jam_mengajar_per_minggu::double precision,
	       pb.status_di_kurikulum::integer, (pb.induk_pembelajaran_id IS NULL), pb.induk_pembelajaran_id::text
	FROM public.pembelajaran AS pb
	JOIN public.ptk_terdaftar AS terdaftar ON terdaftar.ptk_terdaftar_id = pb.ptk_terdaftar_id AND terdaftar.soft_delete = 0
	JOIN public.ptk AS ptk ON ptk.ptk_id = terdaftar.ptk_id AND ptk.soft_delete = 0
	LEFT JOIN ref.mata_pelajaran AS mp ON mp.mata_pelajaran_id = pb.mata_pelajaran_id AND mp.expired_date IS NULL
	WHERE pb.rombongan_belajar_id = $1::uuid AND pb.soft_delete = 0
	  AND ($2::text = '' OR btrim(pb.semester_id) = $2::text)
	ORDER BY COALESCE(mp.nama, pb.nama_mata_pelajaran), ptk.nama, pb.pembelajaran_id`
	rows, err := r.pool.Query(ctx, query, id, r.period.SemesterID)
	if err != nil {
		return nil, fmt.Errorf("query pembelajaran rombel: %w", err)
	}
	defer rows.Close()
	items := make([]Pembelajaran, 0)
	for rows.Next() {
		var item Pembelajaran
		if err := rows.Scan(&item.ID, &item.SemesterID, &item.MataPelajaranID, &item.KodeMataPelajaran, &item.MataPelajaran, &item.PTKID, &item.NamaGuru, &item.PTKTerdaftarID, &item.SKMengajar, &item.TanggalSK, &item.JamPerMinggu, &item.StatusDiKurikulum, &item.Induk, &item.IndukPembelajaranID); err != nil {
			return nil, fmt.Errorf("scan pembelajaran rombel: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate pembelajaran rombel: %w", err)
	}
	return items, nil
}

func (r *PostgresRepository) ListAllPembelajaran(ctx context.Context) ([]Pembelajaran, error) {
	const query = `
	SELECT pb.pembelajaran_id::text, pb.rombongan_belajar_id::text, rb.nama, wali.nama,
	       btrim(pb.semester_id), pb.mata_pelajaran_id, NULL::text,
	       COALESCE(mp.nama, pb.nama_mata_pelajaran), ptk.ptk_id::text, ptk.nama, terdaftar.ptk_terdaftar_id::text,
	       pb.sk_mengajar, to_char(pb.tanggal_sk_mengajar, 'YYYY-MM-DD'), pb.jam_mengajar_per_minggu::double precision,
	       pb.status_di_kurikulum::integer, (pb.induk_pembelajaran_id IS NULL), pb.induk_pembelajaran_id::text
	FROM public.pembelajaran AS pb
	JOIN public.rombongan_belajar AS rb ON rb.rombongan_belajar_id = pb.rombongan_belajar_id AND rb.soft_delete = 0
	LEFT JOIN public.ptk AS wali ON wali.ptk_id = rb.ptk_id AND wali.soft_delete = 0
	JOIN public.ptk_terdaftar AS terdaftar ON terdaftar.ptk_terdaftar_id = pb.ptk_terdaftar_id AND terdaftar.soft_delete = 0
	JOIN public.ptk AS ptk ON ptk.ptk_id = terdaftar.ptk_id AND ptk.soft_delete = 0
	LEFT JOIN ref.mata_pelajaran AS mp ON mp.mata_pelajaran_id = pb.mata_pelajaran_id AND mp.expired_date IS NULL
	WHERE pb.soft_delete = 0 AND ($1::text = '' OR btrim(pb.semester_id) = $1::text)
	ORDER BY rb.nama, COALESCE(mp.nama, pb.nama_mata_pelajaran), ptk.nama, pb.pembelajaran_id`
	rows, err := r.pool.Query(ctx, query, r.period.SemesterID)
	if err != nil {
		return nil, fmt.Errorf("query seluruh pembelajaran: %w", err)
	}
	defer rows.Close()
	items := make([]Pembelajaran, 0)
	for rows.Next() {
		var item Pembelajaran
		if err := rows.Scan(&item.ID, &item.RombelID, &item.NamaKelas, &item.WaliKelas, &item.SemesterID, &item.MataPelajaranID, &item.KodeMataPelajaran, &item.MataPelajaran, &item.PTKID, &item.NamaGuru, &item.PTKTerdaftarID, &item.SKMengajar, &item.TanggalSK, &item.JamPerMinggu, &item.StatusDiKurikulum, &item.Induk, &item.IndukPembelajaranID); err != nil {
			return nil, fmt.Errorf("scan seluruh pembelajaran: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate seluruh pembelajaran: %w", err)
	}
	return items, nil
}

func (r *PostgresRepository) GetStatistics(ctx context.Context, id string) (Statistics, error) {
	result := Statistics{Agama: make([]Distribution, 0), Usia: make([]AgeDistribution, 0), EvaluasiRaporTersedia: false}
	const base = `
	FROM public.anggota_rombel AS ar
	JOIN public.peserta_didik AS pd ON pd.peserta_didik_id = ar.peserta_didik_id AND pd.soft_delete = 0
	JOIN public.rombongan_belajar AS rb ON rb.rombongan_belajar_id = ar.rombongan_belajar_id AND rb.soft_delete = 0
	LEFT JOIN ref.agama AS agama ON agama.agama_id = pd.agama_id AND agama.expired_date IS NULL
	WHERE ar.rombongan_belajar_id = $1::uuid AND ar.soft_delete = 0
	  AND ($2::text = '' OR btrim(rb.semester_id) = $2::text)`
	if err := r.pool.QueryRow(ctx, `SELECT count(DISTINCT pd.peserta_didik_id)::bigint, count(DISTINCT pd.peserta_didik_id) FILTER (WHERE btrim(pd.jenis_kelamin) = 'L')::bigint, count(DISTINCT pd.peserta_didik_id) FILTER (WHERE btrim(pd.jenis_kelamin) = 'P')::bigint, count(DISTINCT pd.peserta_didik_id) FILTER (WHERE btrim(pd.jenis_kelamin) NOT IN ('L','P') OR pd.jenis_kelamin IS NULL)::bigint `+base, id, r.period.SemesterID).Scan(&result.TotalSiswa, &result.LakiLaki, &result.Perempuan, &result.TidakDiketahui); err != nil {
		return result, fmt.Errorf("count statistik rombel: %w", err)
	}
	rows, err := r.pool.Query(ctx, `SELECT COALESCE(pd.agama_id::integer, 0), COALESCE(agama.nama, 'Tidak diketahui'), count(DISTINCT pd.peserta_didik_id)::bigint `+base+` GROUP BY pd.agama_id, agama.nama ORDER BY pd.agama_id NULLS LAST`, id, r.period.SemesterID)
	if err != nil {
		return result, fmt.Errorf("query sebaran agama rombel: %w", err)
	}
	for rows.Next() {
		var item Distribution
		if err := rows.Scan(&item.ID, &item.Label, &item.Jumlah); err != nil {
			rows.Close()
			return result, fmt.Errorf("scan sebaran agama: %w", err)
		}
		result.Agama = append(result.Agama, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return result, fmt.Errorf("iterate sebaran agama: %w", err)
	}
	rows.Close()
	rows, err = r.pool.Query(ctx, `SELECT CASE WHEN pd.tanggal_lahir IS NULL THEN 'tidak_diketahui' WHEN date_part('year', age(current_date, pd.tanggal_lahir)) < 6 THEN '<6' WHEN date_part('year', age(current_date, pd.tanggal_lahir)) BETWEEN 6 AND 11 THEN '6-11' WHEN date_part('year', age(current_date, pd.tanggal_lahir)) BETWEEN 12 AND 15 THEN '12-15' WHEN date_part('year', age(current_date, pd.tanggal_lahir)) BETWEEN 16 AND 18 THEN '16-18' ELSE '>18' END, count(DISTINCT pd.peserta_didik_id)::bigint `+base+` GROUP BY 1 ORDER BY 1`, id, r.period.SemesterID)
	if err != nil {
		return result, fmt.Errorf("query sebaran usia rombel: %w", err)
	}
	for rows.Next() {
		var item AgeDistribution
		if err := rows.Scan(&item.Rentang, &item.Jumlah); err != nil {
			rows.Close()
			return result, fmt.Errorf("scan sebaran usia: %w", err)
		}
		result.Usia = append(result.Usia, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return result, fmt.Errorf("iterate sebaran usia: %w", err)
	}
	rows.Close()
	if err := r.pool.QueryRow(ctx, `SELECT count(*)::bigint, count(DISTINCT mata_pelajaran_id)::bigint FROM public.pembelajaran WHERE rombongan_belajar_id = $1::uuid AND soft_delete = 0 AND ($2::text = '' OR btrim(semester_id) = $2::text)`, id, r.period.SemesterID).Scan(&result.JumlahPembelajaran, &result.JumlahMataPelajaran); err != nil {
		return result, fmt.Errorf("count pembelajaran rombel: %w", err)
	}
	return result, nil
}
