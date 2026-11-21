package pesertadidik

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("peserta didik tidak ditemukan")

type Store interface {
	List(context.Context, int, int, Filters) ([]ListItem, int64, error)
	GetDetail(context.Context, string) (Detail, error)
	ListRombel(context.Context, string) ([]Rombel, error)
	GetPrestasiBeasiswa(context.Context, string) (PrestasiBeasiswa, error)
}

type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

const currentRegistration = `
WITH registration AS (
	SELECT DISTINCT ON (r.peserta_didik_id)
	       r.peserta_didik_id, r.registrasi_id, r.sekolah_id, r.nipd,
	       r.tanggal_masuk_sekolah, r.jenis_pendaftaran_id, r.jenis_keluar_id,
	       r.sekolah_asal, r.last_update
	FROM public.registrasi_peserta_didik AS r
	WHERE r.soft_delete = 0
	ORDER BY r.peserta_didik_id, r.last_update DESC, r.registrasi_id DESC
)`

const studentListFrom = `
FROM public.peserta_didik AS pd
JOIN registration AS reg ON reg.peserta_didik_id = pd.peserta_didik_id
LEFT JOIN LATERAL (
	SELECT ar.rombongan_belajar_id, rb.nama, rb.semester_id, rb.tingkat_pendidikan_id
	FROM public.anggota_rombel AS ar
	JOIN public.rombongan_belajar AS rb ON rb.rombongan_belajar_id = ar.rombongan_belajar_id AND rb.soft_delete = 0
	WHERE ar.peserta_didik_id = pd.peserta_didik_id AND ar.soft_delete = 0
	ORDER BY rb.semester_id DESC, ar.last_update DESC
	LIMIT 1
) AS kelas ON TRUE
LEFT JOIN ref.tingkat_pendidikan AS tingkat ON tingkat.tingkat_pendidikan_id = kelas.tingkat_pendidikan_id AND tingkat.expired_date IS NULL
WHERE pd.soft_delete = 0
	AND ($1::uuid IS NULL OR reg.sekolah_id = $1::uuid)
	AND ($2::uuid IS NULL OR EXISTS (
		SELECT 1 FROM public.anggota_rombel AS ar
		JOIN public.rombongan_belajar AS rb ON rb.rombongan_belajar_id = ar.rombongan_belajar_id AND rb.soft_delete = 0
		WHERE ar.peserta_didik_id = pd.peserta_didik_id AND ar.rombongan_belajar_id = $2::uuid AND ar.soft_delete = 0
	))
	AND ($3::text = '' OR EXISTS (
		SELECT 1 FROM public.anggota_rombel AS ar
		JOIN public.rombongan_belajar AS rb ON rb.rombongan_belajar_id = ar.rombongan_belajar_id AND rb.soft_delete = 0
		WHERE ar.peserta_didik_id = pd.peserta_didik_id AND btrim(rb.semester_id) = $3::text AND ar.soft_delete = 0
	))
	AND ($4::bigint IS NULL OR EXISTS (
		SELECT 1 FROM public.anggota_rombel AS ar
		JOIN public.rombongan_belajar AS rb ON rb.rombongan_belajar_id = ar.rombongan_belajar_id AND rb.soft_delete = 0
		WHERE ar.peserta_didik_id = pd.peserta_didik_id AND rb.tingkat_pendidikan_id = $4::bigint AND ar.soft_delete = 0
	))
	AND ($5::boolean IS NULL OR (($5::boolean AND reg.jenis_keluar_id IS NULL) OR (NOT $5::boolean AND reg.jenis_keluar_id IS NOT NULL)))
	AND ($6::text = '' OR pd.nama ILIKE '%' || $6::text || '%' OR btrim(pd.nisn) ILIKE '%' || $6::text || '%' OR reg.nipd ILIKE '%' || $6::text || '%')`

func (r *PostgresRepository) List(ctx context.Context, page, limit int, f Filters) ([]ListItem, int64, error) {
	args := filterArgs(f)
	var total int64
	if err := r.pool.QueryRow(ctx, currentRegistration+` SELECT count(DISTINCT pd.peserta_didik_id) `+studentListFrom, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count peserta didik: %w", err)
	}
	query := currentRegistration + `
	SELECT pd.peserta_didik_id::text, pd.nama, btrim(pd.jenis_kelamin), NULLIF(btrim(pd.nisn), ''),
	       NULLIF(reg.nipd, ''), reg.sekolah_id::text, reg.jenis_keluar_id,
	       kelas.rombongan_belajar_id::text, kelas.nama, NULLIF(btrim(kelas.semester_id), ''),
	       kelas.tingkat_pendidikan_id::bigint, tingkat.nama ` + studentListFrom + `
	ORDER BY pd.nama, pd.peserta_didik_id`
	if limit > 0 {
		query += ` LIMIT $7 OFFSET $8`
		args = append(args, limit, (page-1)*limit)
	}
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query peserta didik: %w", err)
	}
	defer rows.Close()
	items := make([]ListItem, 0)
	for rows.Next() {
		var item ListItem
		if err := rows.Scan(&item.PesertaDidikID, &item.Nama, &item.JenisKelamin, &item.NISN, &item.NIPD, &item.SekolahID, &item.JenisKeluarID, &item.RombelID, &item.Rombel, &item.SemesterID, &item.TingkatID, &item.Tingkat); err != nil {
			return nil, 0, fmt.Errorf("scan peserta didik: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate peserta didik: %w", err)
	}
	return items, total, nil
}

func filterArgs(f Filters) []any {
	var aktif any
	if f.Aktif != nil {
		aktif = *f.Aktif
	}
	return []any{nullable(f.SekolahID), nullable(f.RombonganBelajarID), f.SemesterID, nullable(f.TingkatPendidikanID), aktif, f.Search}
}

func nullable(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func (r *PostgresRepository) GetDetail(ctx context.Context, id string) (Detail, error) {
	const query = `
	SELECT pd.peserta_didik_id::text, pd.nama, btrim(pd.jenis_kelamin), NULLIF(btrim(pd.nisn), ''),
	       pd.tempat_lahir, to_char(pd.tanggal_lahir, 'YYYY-MM-DD'), pd.agama_id::integer, agama.nama,
	       pd.alamat_jalan, pd.rt::bigint, pd.rw::bigint, pd.nama_dusun, pd.desa_kelurahan,
	       btrim(pd.kode_wilayah), NULLIF(btrim(pd.kode_pos), ''), pd.jenis_tinggal_id::bigint, jt.nama,
	       pd.alat_transportasi_id::bigint, transportasi.nama, pd.nomor_telepon_seluler, pd.email,
	       pd.anak_keberapa::bigint, pd.pekerjaan_id, pekerjaan.nama,
	       reg.sekolah_id::text, reg.nipd, to_char(reg.tanggal_masuk_sekolah, 'YYYY-MM-DD'),
	       reg.jenis_pendaftaran_id::bigint, pendaftaran.nama, reg.jenis_keluar_id, keluar.ket_keluar, reg.sekolah_asal,
	       pd.nama_ayah, pd.tahun_lahir_ayah::bigint, pd.jenjang_pendidikan_ayah::bigint,
	       pd.pekerjaan_id_ayah, pa.nama, pd.penghasilan_id_ayah, pga.nama,
	       pd.nama_ibu_kandung, pd.tahun_lahir_ibu::bigint, pd.jenjang_pendidikan_ibu::bigint,
	       pd.pekerjaan_id_ibu, pi.nama, pd.penghasilan_id_ibu, pgi.nama,
	       pd.nama_wali, pd.tahun_lahir_wali::bigint, pd.jenjang_pendidikan_wali::bigint,
	       pd.pekerjaan_id_wali, pw.nama, pd.penghasilan_id_wali, pgw.nama,
	       longitudinal.tinggi_badan::double precision, longitudinal.berat_badan::double precision,
	       longitudinal.jarak_rumah_ke_sekolah_km::double precision, NULLIF(btrim(longitudinal.semester_id), '')
	FROM public.peserta_didik AS pd
	LEFT JOIN ref.agama AS agama ON agama.agama_id = pd.agama_id AND agama.expired_date IS NULL
	LEFT JOIN ref.jenis_tinggal AS jt ON jt.jenis_tinggal_id = pd.jenis_tinggal_id AND jt.expired_date IS NULL
	LEFT JOIN ref.alat_transportasi AS transportasi ON transportasi.alat_transportasi_id = pd.alat_transportasi_id AND transportasi.expired_date IS NULL
	LEFT JOIN ref.pekerjaan AS pekerjaan ON pekerjaan.pekerjaan_id = pd.pekerjaan_id AND pekerjaan.expired_date IS NULL
	LEFT JOIN LATERAL (SELECT r.peserta_didik_id, r.registrasi_id, r.sekolah_id, r.nipd, r.tanggal_masuk_sekolah, r.jenis_pendaftaran_id, r.jenis_keluar_id, r.sekolah_asal, r.last_update FROM public.registrasi_peserta_didik AS r WHERE r.peserta_didik_id = pd.peserta_didik_id AND r.soft_delete = 0 ORDER BY r.last_update DESC, r.registrasi_id DESC LIMIT 1) AS reg ON TRUE
	LEFT JOIN ref.jenis_pendaftaran AS pendaftaran ON pendaftaran.jenis_pendaftaran_id = reg.jenis_pendaftaran_id AND pendaftaran.expired_date IS NULL
	LEFT JOIN ref.jenis_keluar AS keluar ON keluar.jenis_keluar_id = reg.jenis_keluar_id AND keluar.expired_date IS NULL
	LEFT JOIN ref.pekerjaan AS pa ON pa.pekerjaan_id = pd.pekerjaan_id_ayah AND pa.expired_date IS NULL
	LEFT JOIN ref.penghasilan AS pga ON pga.penghasilan_id = pd.penghasilan_id_ayah AND pga.expired_date IS NULL
	LEFT JOIN ref.pekerjaan AS pi ON pi.pekerjaan_id = pd.pekerjaan_id_ibu AND pi.expired_date IS NULL
	LEFT JOIN ref.penghasilan AS pgi ON pgi.penghasilan_id = pd.penghasilan_id_ibu AND pgi.expired_date IS NULL
	LEFT JOIN ref.pekerjaan AS pw ON pw.pekerjaan_id = pd.pekerjaan_id_wali AND pw.expired_date IS NULL
	LEFT JOIN ref.penghasilan AS pgw ON pgw.penghasilan_id = pd.penghasilan_id_wali AND pgw.expired_date IS NULL
	LEFT JOIN LATERAL (SELECT l.tinggi_badan, l.berat_badan, l.jarak_rumah_ke_sekolah_km, l.semester_id FROM public.peserta_didik_longitudinal AS l WHERE l.peserta_didik_id = pd.peserta_didik_id AND l.soft_delete = 0 ORDER BY l.semester_id DESC LIMIT 1) AS longitudinal ON TRUE
	WHERE pd.peserta_didik_id = $1::uuid AND pd.soft_delete = 0`
	var item Detail
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&item.PesertaDidikID, &item.Nama, &item.JenisKelamin, &item.NISN, &item.TempatLahir, &item.TanggalLahir, &item.AgamaID, &item.Agama,
		&item.AlamatJalan, &item.RT, &item.RW, &item.NamaDusun, &item.DesaKelurahan, &item.KodeWilayah, &item.KodePos, &item.JenisTinggalID, &item.JenisTinggal,
		&item.AlatTransportasiID, &item.AlatTransportasi, &item.NomorTeleponSeluler, &item.Email, &item.AnakKeberapa, &item.PekerjaanID, &item.Pekerjaan,
		&item.SekolahID, &item.NIPD, &item.TanggalMasuk, &item.JenisPendaftaranID, &item.JenisPendaftaran, &item.JenisKeluarID, &item.JenisKeluar, &item.SekolahAsal,
		&item.NamaAyah, &item.TahunLahirAyah, &item.PendidikanAyahID, &item.PekerjaanAyahID, &item.PekerjaanAyah, &item.PenghasilanAyahID, &item.PenghasilanAyah,
		&item.NamaIbu, &item.TahunLahirIbu, &item.PendidikanIbuID, &item.PekerjaanIbuID, &item.PekerjaanIbu, &item.PenghasilanIbuID, &item.PenghasilanIbu,
		&item.NamaWali, &item.TahunLahirWali, &item.PendidikanWaliID, &item.PekerjaanWaliID, &item.PekerjaanWali, &item.PenghasilanWaliID, &item.PenghasilanWali,
		&item.TinggiBadan, &item.BeratBadan, &item.JarakRumahSekolahKM, &item.SemesterLongitudinal)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Detail{}, ErrNotFound
		}
		return Detail{}, fmt.Errorf("query detail peserta didik: %w", err)
	}
	return item, nil
}

func (r *PostgresRepository) ListRombel(ctx context.Context, id string) ([]Rombel, error) {
	const query = `
	SELECT ar.anggota_rombel_id::text, rb.rombongan_belajar_id::text, rb.nama, btrim(rb.semester_id), semester.nama,
	       rb.sekolah_id::text, rb.tingkat_pendidikan_id::bigint, tingkat.nama, rb.ptk_id::text, ptk.nama,
	       ar.jenis_pendaftaran_id::bigint, pendaftaran.nama
	FROM public.anggota_rombel AS ar
	JOIN public.rombongan_belajar AS rb ON rb.rombongan_belajar_id = ar.rombongan_belajar_id AND rb.soft_delete = 0
	LEFT JOIN ref.semester AS semester ON btrim(semester.semester_id) = btrim(rb.semester_id) AND semester.expired_date IS NULL
	LEFT JOIN ref.tingkat_pendidikan AS tingkat ON tingkat.tingkat_pendidikan_id = rb.tingkat_pendidikan_id AND tingkat.expired_date IS NULL
	LEFT JOIN public.ptk AS ptk ON ptk.ptk_id = rb.ptk_id AND ptk.soft_delete = 0
	LEFT JOIN ref.jenis_pendaftaran AS pendaftaran ON pendaftaran.jenis_pendaftaran_id = ar.jenis_pendaftaran_id AND pendaftaran.expired_date IS NULL
	WHERE ar.peserta_didik_id = $1::uuid AND ar.soft_delete = 0
	ORDER BY rb.semester_id DESC, rb.nama`
	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("query riwayat rombel peserta didik: %w", err)
	}
	defer rows.Close()
	items := make([]Rombel, 0)
	for rows.Next() {
		var item Rombel
		if err := rows.Scan(&item.AnggotaRombelID, &item.RombelID, &item.Nama, &item.SemesterID, &item.Semester, &item.SekolahID, &item.TingkatID, &item.Tingkat, &item.WaliKelasID, &item.WaliKelas, &item.JenisPendaftaranID, &item.JenisPendaftaran); err != nil {
			return nil, fmt.Errorf("scan riwayat rombel: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate riwayat rombel: %w", err)
	}
	return items, nil
}

func (r *PostgresRepository) GetPrestasiBeasiswa(ctx context.Context, id string) (PrestasiBeasiswa, error) {
	result := PrestasiBeasiswa{Prestasi: make([]Prestasi, 0), Beasiswa: make([]Beasiswa, 0)}
	rows, err := r.pool.Query(ctx, `SELECT p.prestasi_id::text, p.jenis_prestasi_id, jp.nama, p.tingkat_prestasi_id, tp.nama, p.nama, p.tahun_prestasi::bigint, p.penyelenggara, p.peringkat FROM public.prestasi AS p LEFT JOIN ref.jenis_prestasi AS jp ON jp.jenis_prestasi_id = p.jenis_prestasi_id AND jp.expired_date IS NULL LEFT JOIN ref.tingkat_prestasi AS tp ON tp.tingkat_prestasi_id = p.tingkat_prestasi_id AND tp.expired_date IS NULL WHERE p.peserta_didik_id = $1::uuid AND p.soft_delete = 0 ORDER BY p.tahun_prestasi DESC, p.nama`, id)
	if err != nil {
		return result, fmt.Errorf("query prestasi peserta didik: %w", err)
	}
	for rows.Next() {
		var item Prestasi
		if err := rows.Scan(&item.ID, &item.JenisPrestasiID, &item.JenisPrestasi, &item.TingkatPrestasiID, &item.TingkatPrestasi, &item.Nama, &item.Tahun, &item.Penyelenggara, &item.Peringkat); err != nil {
			rows.Close()
			return result, fmt.Errorf("scan prestasi: %w", err)
		}
		result.Prestasi = append(result.Prestasi, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return result, fmt.Errorf("iterate prestasi: %w", err)
	}
	rows.Close()
	rows, err = r.pool.Query(ctx, `SELECT b.beasiswa_peserta_didik_id::text, b.jenis_beasiswa_id, jb.nama, b.keterangan, b.tahun_mulai::bigint, b.tahun_selesai::bigint FROM public.beasiswa_peserta_didik AS b LEFT JOIN ref.jenis_beasiswa AS jb ON jb.jenis_beasiswa_id = b.jenis_beasiswa_id AND jb.expired_date IS NULL WHERE b.peserta_didik_id = $1::uuid AND b.soft_delete = 0 ORDER BY b.tahun_mulai DESC, b.jenis_beasiswa_id`, id)
	if err != nil {
		return result, fmt.Errorf("query beasiswa peserta didik: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item Beasiswa
		if err := rows.Scan(&item.ID, &item.JenisBeasiswaID, &item.JenisBeasiswa, &item.Keterangan, &item.TahunMulai, &item.TahunSelesai); err != nil {
			return result, fmt.Errorf("scan beasiswa: %w", err)
		}
		result.Beasiswa = append(result.Beasiswa, item)
	}
	if err := rows.Err(); err != nil {
		return result, fmt.Errorf("iterate beasiswa: %w", err)
	}
	return result, nil
}
