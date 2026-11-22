package ptk

import (
	"context"
	"errors"
	"fmt"

	"dapodik-readonly-api/internal/period"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("ptk tidak ditemukan")

type Store interface {
	List(context.Context, int, int, Filters) ([]ListItem, int64, error)
	GetDetail(context.Context, string) (Detail, error)
	ListPembelajaran(context.Context, string, string) ([]Pembelajaran, error)
	ListTugasTambahan(context.Context, string) ([]TugasTambahan, error)
	GetRiwayat(context.Context, string) (Riwayat, error)
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

const activeRegistration = `
	WITH registration AS (
		SELECT DISTINCT ON (ptk_id)
		       ptk_id,
		       sekolah_id,
		       tahun_ajaran_id,
		       jenis_ptk_id,
		       jabatan_ptk_id
		FROM public.ptk_terdaftar
		WHERE soft_delete = 0
		  AND ($1::text = '' OR sekolah_id::text = $1::text)
		  AND ($2::text = '' OR tahun_ajaran_id::text = $2::text)
		  AND ($3::text = '' OR jenis_ptk_id::text = $3::text)
		ORDER BY ptk_id, tahun_ajaran_id DESC, last_update DESC
	)`

func (r *PostgresRepository) List(ctx context.Context, page, limit int, filters Filters) ([]ListItem, int64, error) {
	if r.period.TahunAjaranID != 0 {
		filters.TahunAjaranID = fmt.Sprint(r.period.TahunAjaranID)
	}
	const searchFilter = `
		AND ($4::text = ''
		     OR p.nama ILIKE '%' || $4::text || '%'
		     OR btrim(p.nuptk) ILIKE '%' || $4::text || '%'
		     OR p.nip ILIKE '%' || $4::text || '%')`

	var total int64
	countQuery := activeRegistration + `
		SELECT count(*)
		FROM public.ptk AS p
		JOIN registration AS r ON r.ptk_id = p.ptk_id
		WHERE p.soft_delete = 0` + searchFilter
	if err := r.pool.QueryRow(ctx, countQuery, filters.SekolahID, filters.TahunAjaranID, filters.JenisPTKID, filters.Search).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count ptk: %w", err)
	}

	listQuery := activeRegistration + `
		SELECT p.ptk_id::text,
		       p.nama,
		       p.nip,
		       NULLIF(btrim(p.nuptk), ''),
		       p.nrg,
	       btrim(p.nik),
	       NULLIF(btrim(p.no_kk), ''),
	       NULLIF(btrim(p.npwp), ''),
	       NULLIF(btrim(p.rekening_bank), ''),
	       NULLIF(btrim(p.rekening_atas_nama), ''),
	       NULLIF(btrim(p.karpeg), ''),
	       NULLIF(btrim(p.karpas), ''),
		       btrim(p.jenis_kelamin),
		       p.status_kepegawaian_id::integer,
		       sk.nama,
		       r.sekolah_id::text,
	       r.tahun_ajaran_id::bigint,
	       r.jenis_ptk_id::bigint,
	       jp.jenis_ptk
		FROM public.ptk AS p
		JOIN registration AS r ON r.ptk_id = p.ptk_id
		LEFT JOIN ref.status_kepegawaian AS sk
		  ON sk.status_kepegawaian_id = p.status_kepegawaian_id
		 AND sk.expired_date IS NULL
		LEFT JOIN ref.jenis_ptk AS jp
		  ON jp.jenis_ptk_id = r.jenis_ptk_id
		 AND jp.expired_date IS NULL
		WHERE p.soft_delete = 0` + searchFilter + `
		ORDER BY p.nama, p.ptk_id`
	args := []any{filters.SekolahID, filters.TahunAjaranID, filters.JenisPTKID, filters.Search}
	if limit > 0 {
		listQuery += ` LIMIT $5 OFFSET $6`
		args = append(args, limit, (page-1)*limit)
	}
	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query ptk: %w", err)
	}
	defer rows.Close()

	items := make([]ListItem, 0)
	for rows.Next() {
		var item ListItem
		if err := rows.Scan(
			&item.PTKID,
			&item.Nama,
			&item.NIP,
			&item.NUPTK,
			&item.NRG,
			&item.NIK,
			&item.NoKK,
			&item.NPWP,
			&item.RekeningBank,
			&item.RekeningAtasNama,
			&item.Karpeg,
			&item.Karpas,
			&item.JenisKelamin,
			&item.StatusKepegawaianID,
			&item.StatusKepegawaian,
			&item.SekolahID,
			&item.TahunAjaranID,
			&item.JenisPTKID,
			&item.JenisPTK,
		); err != nil {
			return nil, 0, fmt.Errorf("scan ptk: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate ptk: %w", err)
	}
	return items, total, nil
}

func (r *PostgresRepository) GetDetail(ctx context.Context, id string) (Detail, error) {
	const query = `
		SELECT p.ptk_id::text,
	       p.nama,
	       p.nip,
	       NULLIF(btrim(p.nuptk), ''),
	       p.nrg,
	       p.nuks,
	       btrim(p.nik),
	       NULLIF(btrim(p.no_kk), ''),
	       NULLIF(btrim(p.npwp), ''),
	       NULLIF(btrim(p.rekening_bank), ''),
	       NULLIF(btrim(p.rekening_atas_nama), ''),
	       NULLIF(btrim(p.karpeg), ''),
	       NULLIF(btrim(p.karpas), ''),
	       btrim(p.jenis_kelamin),
	       p.tempat_lahir,
	       to_char(p.tanggal_lahir, 'YYYY-MM-DD'),
	       p.agama_id::integer,
	       agama.nama,
	       p.status_kepegawaian_id::integer,
	       sk.nama,
	       p.status_keaktifan_id::bigint,
	       sa.nama,
	       p.pangkat_golongan_id::bigint,
	       pg.nama,
	       p.alamat_jalan,
	       p.rt::bigint,
	       p.rw::bigint,
	       p.nama_dusun,
	       p.desa_kelurahan,
	       btrim(p.kode_wilayah),
	       wilayah.nama,
	       NULLIF(btrim(p.kode_pos), ''),
	       p.lintang::double precision,
	       p.bujur::double precision,
	       p.no_hp,
	       p.email,
	       registration.sekolah_id::text,
	       registration.tahun_ajaran_id::bigint,
	       registration.jenis_ptk_id::bigint,
	       jenis.jenis_ptk,
	       registration.jabatan_ptk_id::bigint,
	       jabatan.jabatan_ptk
	FROM public.ptk AS p
	LEFT JOIN ref.agama AS agama
	  ON agama.agama_id = p.agama_id
	 AND agama.expired_date IS NULL
	LEFT JOIN ref.status_kepegawaian AS sk
	  ON sk.status_kepegawaian_id = p.status_kepegawaian_id
	 AND sk.expired_date IS NULL
	LEFT JOIN ref.status_keaktifan_pegawai AS sa
	  ON sa.status_keaktifan_id = p.status_keaktifan_id
	 AND sa.expired_date IS NULL
	LEFT JOIN ref.pangkat_golongan AS pg
	  ON pg.pangkat_golongan_id = p.pangkat_golongan_id
	 AND pg.expired_date IS NULL
	LEFT JOIN ref.mst_wilayah AS wilayah
	  ON wilayah.kode_wilayah = p.kode_wilayah
	 AND wilayah.expired_date IS NULL
	LEFT JOIN (
		SELECT DISTINCT ON (pt.ptk_id)
		       pt.ptk_id, pt.sekolah_id, pt.tahun_ajaran_id,
		       pt.jenis_ptk_id, pt.jabatan_ptk_id
		FROM public.ptk_terdaftar AS pt
		WHERE pt.soft_delete = 0
		  AND ($2::bigint = 0 OR pt.tahun_ajaran_id = $2::bigint)
		ORDER BY pt.ptk_id, pt.tahun_ajaran_id DESC, pt.last_update DESC
	) AS registration ON registration.ptk_id = p.ptk_id
	LEFT JOIN ref.jenis_ptk AS jenis
	  ON jenis.jenis_ptk_id = registration.jenis_ptk_id
	 AND jenis.expired_date IS NULL
	LEFT JOIN ref.jabatan_ptk AS jabatan
	  ON jabatan.jabatan_ptk_id = registration.jabatan_ptk_id
	 AND jabatan.expired_date IS NULL
	WHERE p.ptk_id = $1::uuid
	  AND p.soft_delete = 0`

	var item Detail
	err := r.pool.QueryRow(ctx, query, id, r.period.TahunAjaranID).Scan(
		&item.PTKID,
		&item.Nama,
		&item.NIP,
		&item.NUPTK,
		&item.NRG,
		&item.NUKS,
		&item.NIK,
		&item.NoKK,
		&item.NPWP,
		&item.RekeningBank,
		&item.RekeningAtasNama,
		&item.Karpeg,
		&item.Karpas,
		&item.JenisKelamin,
		&item.TempatLahir,
		&item.TanggalLahir,
		&item.AgamaID,
		&item.Agama,
		&item.StatusKepegawaianID,
		&item.StatusKepegawaian,
		&item.StatusKeaktifanID,
		&item.StatusKeaktifan,
		&item.PangkatGolonganID,
		&item.PangkatGolongan,
		&item.AlamatJalan,
		&item.RT,
		&item.RW,
		&item.NamaDusun,
		&item.DesaKelurahan,
		&item.KodeWilayah,
		&item.NamaWilayah,
		&item.KodePos,
		&item.Lintang,
		&item.Bujur,
		&item.NoHP,
		&item.Email,
		&item.SekolahID,
		&item.TahunAjaranID,
		&item.JenisPTKID,
		&item.JenisPTK,
		&item.JabatanPTKID,
		&item.JabatanPTK,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Detail{}, ErrNotFound
		}
		return Detail{}, fmt.Errorf("query ptk detail: %w", err)
	}
	return item, nil
}

func (r *PostgresRepository) ListPembelajaran(ctx context.Context, id, semesterID string) ([]Pembelajaran, error) {
	if r.period.SemesterID != "" {
		semesterID = r.period.SemesterID
	}
	const query = `
		SELECT pb.pembelajaran_id::text,
	       btrim(pb.semester_id),
	       pb.rombongan_belajar_id::text,
	       rb.nama,
	       pb.mata_pelajaran_id,
	       COALESCE(mp.nama, pb.nama_mata_pelajaran),
	       pb.sk_mengajar,
	       to_char(pb.tanggal_sk_mengajar, 'YYYY-MM-DD'),
	       pb.jam_mengajar_per_minggu::double precision,
	       pb.status_di_kurikulum::integer
	FROM public.pembelajaran AS pb
	JOIN public.ptk_terdaftar AS terdaftar
	  ON terdaftar.ptk_terdaftar_id = pb.ptk_terdaftar_id
	 AND terdaftar.ptk_id = $1::uuid
	 AND terdaftar.soft_delete = 0
	LEFT JOIN public.rombongan_belajar AS rb
	  ON rb.rombongan_belajar_id = pb.rombongan_belajar_id
	 AND rb.soft_delete = 0
	LEFT JOIN ref.mata_pelajaran AS mp
	  ON mp.mata_pelajaran_id = pb.mata_pelajaran_id
	 AND mp.expired_date IS NULL
	WHERE pb.soft_delete = 0
	  AND ($2::text = '' OR btrim(pb.semester_id) = $2::text)
	ORDER BY pb.semester_id DESC, rb.nama, mp.nama`

	rows, err := r.pool.Query(ctx, query, id, semesterID)
	if err != nil {
		return nil, fmt.Errorf("query ptk pembelajaran: %w", err)
	}
	defer rows.Close()

	items := make([]Pembelajaran, 0)
	for rows.Next() {
		var item Pembelajaran
		if err := rows.Scan(&item.ID, &item.SemesterID, &item.RombonganBelajarID, &item.Rombel, &item.MataPelajaranID, &item.MataPelajaran, &item.SKMengajar, &item.TanggalSKMengajar, &item.JamMengajarPerMinggu, &item.StatusDiKurikulum); err != nil {
			return nil, fmt.Errorf("scan ptk pembelajaran: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ptk pembelajaran: %w", err)
	}
	return items, nil
}

func (r *PostgresRepository) ListTugasTambahan(ctx context.Context, id string) ([]TugasTambahan, error) {
	const query = `
		SELECT tt.ptk_tugas_tambahan_id::text,
	       tt.jabatan_ptk_id::bigint,
	       jabatan.nama,
	       tt.sekolah_id::text,
	       tt.jumlah_jam::double precision,
	       tt.nomor_sk,
	       to_char(tt.tmt_tambahan, 'YYYY-MM-DD'),
	       to_char(tt.tst_tambahan, 'YYYY-MM-DD')
	FROM public.tugas_tambahan AS tt
	LEFT JOIN ref.jabatan_tugas_ptk AS jabatan
	  ON jabatan.jabatan_ptk_id = tt.jabatan_ptk_id
	 AND jabatan.expired_date IS NULL
	WHERE tt.ptk_id = $1::uuid
	  AND tt.soft_delete = 0
	ORDER BY tt.tmt_tambahan DESC, tt.jabatan_ptk_id`

	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("query ptk tugas tambahan: %w", err)
	}
	defer rows.Close()

	items := make([]TugasTambahan, 0)
	for rows.Next() {
		var item TugasTambahan
		if err := rows.Scan(&item.ID, &item.JabatanPTKID, &item.Jabatan, &item.SekolahID, &item.JumlahJam, &item.NomorSK, &item.TMT, &item.TST); err != nil {
			return nil, fmt.Errorf("scan ptk tugas tambahan: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ptk tugas tambahan: %w", err)
	}
	return items, nil
}

func (r *PostgresRepository) GetRiwayat(ctx context.Context, id string) (Riwayat, error) {
	result := Riwayat{
		PendidikanFormal: make([]PendidikanFormal, 0),
		Kepangkatan:      make([]Kepangkatan, 0),
		Sertifikasi:      make([]Sertifikasi, 0),
		Diklat:           make([]Diklat, 0),
	}
	var err error
	if result.PendidikanFormal, err = r.listPendidikan(ctx, id); err != nil {
		return Riwayat{}, err
	}
	if result.Kepangkatan, err = r.listKepangkatan(ctx, id); err != nil {
		return Riwayat{}, err
	}
	if result.Sertifikasi, err = r.listSertifikasi(ctx, id); err != nil {
		return Riwayat{}, err
	}
	if result.Diklat, err = r.listDiklat(ctx, id); err != nil {
		return Riwayat{}, err
	}
	return result, nil
}

func (r *PostgresRepository) listPendidikan(ctx context.Context, id string) ([]PendidikanFormal, error) {
	const query = `
		SELECT rp.riwayat_pendidikan_formal_id::text,
	       rp.jenjang_pendidikan_id::bigint,
	       jenjang.nama,
	       rp.bidang_studi_id,
	       bidang.bidang_studi,
	       rp.gelar_akademik_id,
	       gelar.nama,
	       rp.satuan_pendidikan_formal,
	       rp.fakultas,
	       rp.kependidikan::integer,
	       rp.tahun_masuk::integer,
	       rp.tahun_lulus::integer,
	       rp.status_kuliah::integer,
	       rp.semester::integer,
	       rp.ipk::double precision
	FROM public.rwy_pend_formal AS rp
	LEFT JOIN ref.jenjang_pendidikan AS jenjang
	  ON jenjang.jenjang_pendidikan_id = rp.jenjang_pendidikan_id
	 AND jenjang.expired_date IS NULL
	LEFT JOIN ref.bidang_studi AS bidang
	  ON bidang.bidang_studi_id = rp.bidang_studi_id
	 AND bidang.expired_date IS NULL
	LEFT JOIN ref.gelar_akademik AS gelar
	  ON gelar.gelar_akademik_id = rp.gelar_akademik_id
	 AND gelar.expired_date IS NULL
	WHERE rp.ptk_id = $1::uuid
	  AND rp.soft_delete = 0
	ORDER BY rp.tahun_masuk DESC, rp.riwayat_pendidikan_formal_id`

	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("query ptk pendidikan formal: %w", err)
	}
	defer rows.Close()

	items := make([]PendidikanFormal, 0)
	for rows.Next() {
		var item PendidikanFormal
		if err := rows.Scan(&item.ID, &item.JenjangPendidikanID, &item.JenjangPendidikan, &item.BidangStudiID, &item.BidangStudi, &item.GelarAkademikID, &item.GelarAkademik, &item.SatuanPendidikan, &item.Fakultas, &item.Kependidikan, &item.TahunMasuk, &item.TahunLulus, &item.StatusKuliah, &item.Semester, &item.IPK); err != nil {
			return nil, fmt.Errorf("scan ptk pendidikan formal: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ptk pendidikan formal: %w", err)
	}
	return items, nil
}

func (r *PostgresRepository) listKepangkatan(ctx context.Context, id string) ([]Kepangkatan, error) {
	const query = `
		SELECT rk.riwayat_kepangkatan_id::text,
	       rk.pangkat_golongan_id::bigint,
	       pg.nama,
	       rk.nomor_sk,
	       to_char(rk.tanggal_sk, 'YYYY-MM-DD'),
	       to_char(rk.tmt_pangkat, 'YYYY-MM-DD'),
	       rk.masa_kerja_gol_tahun::integer,
	       rk.masa_kerja_gol_bulan::integer
	FROM public.rwy_kepangkatan AS rk
	LEFT JOIN ref.pangkat_golongan AS pg
	  ON pg.pangkat_golongan_id = rk.pangkat_golongan_id
	 AND pg.expired_date IS NULL
	WHERE rk.ptk_id = $1::uuid
	  AND rk.soft_delete = 0
	ORDER BY rk.tmt_pangkat DESC, rk.riwayat_kepangkatan_id`

	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("query ptk kepangkatan: %w", err)
	}
	defer rows.Close()

	items := make([]Kepangkatan, 0)
	for rows.Next() {
		var item Kepangkatan
		if err := rows.Scan(&item.ID, &item.PangkatGolonganID, &item.PangkatGolongan, &item.NomorSK, &item.TanggalSK, &item.TMTPangkat, &item.MasaKerjaTahun, &item.MasaKerjaBulan); err != nil {
			return nil, fmt.Errorf("scan ptk kepangkatan: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ptk kepangkatan: %w", err)
	}
	return items, nil
}

func (r *PostgresRepository) listSertifikasi(ctx context.Context, id string) ([]Sertifikasi, error) {
	const query = `
		SELECT rs.riwayat_sertifikasi_id::text,
	       rs.bidang_studi_id,
	       bidang.bidang_studi,
	       rs.id_jenis_sertifikasi::bigint,
	       to_char(rs.tgl_sert, 'YYYY-MM-DD'),
	       to_char(rs.tgl_exp_sert, 'YYYY-MM-DD'),
	       rs.nomor_sertifikat,
	       rs.nomer_registrasi,
	       rs.nomor_peserta,
	       rs.kualifikasi
	FROM public.rwy_sertifikasi AS rs
	LEFT JOIN ref.bidang_studi AS bidang
	  ON bidang.bidang_studi_id = rs.bidang_studi_id
	 AND bidang.expired_date IS NULL
	WHERE rs.ptk_id = $1::uuid
	  AND rs.soft_delete = 0
	ORDER BY rs.tgl_sert DESC, rs.riwayat_sertifikasi_id`

	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("query ptk sertifikasi: %w", err)
	}
	defer rows.Close()

	items := make([]Sertifikasi, 0)
	for rows.Next() {
		var item Sertifikasi
		if err := rows.Scan(&item.ID, &item.BidangStudiID, &item.BidangStudi, &item.JenisSertifikasiID, &item.Tanggal, &item.TanggalKedaluwarsa, &item.NomorSertifikat, &item.NomorRegistrasi, &item.NomorPeserta, &item.Kualifikasi); err != nil {
			return nil, fmt.Errorf("scan ptk sertifikasi: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ptk sertifikasi: %w", err)
	}
	return items, nil
}

func (r *PostgresRepository) listDiklat(ctx context.Context, id string) ([]Diklat, error) {
	const query = `
		SELECT diklat_id::text,
	       jenis_diklat_id,
	       nama,
	       penyelenggara,
	       tahun::integer,
	       peran,
	       tingkat::integer,
	       berapa_jam::integer,
	       sertifikat_diklat
	FROM public.diklat
	WHERE ptk_id = $1::uuid
	  AND soft_delete = 0
	ORDER BY tahun DESC, diklat_id`

	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("query ptk diklat: %w", err)
	}
	defer rows.Close()

	items := make([]Diklat, 0)
	for rows.Next() {
		var item Diklat
		if err := rows.Scan(&item.ID, &item.JenisDiklatID, &item.Nama, &item.Penyelenggara, &item.Tahun, &item.Peran, &item.Tingkat, &item.BerapaJam, &item.Sertifikat); err != nil {
			return nil, fmt.Errorf("scan ptk diklat: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ptk diklat: %w", err)
	}
	return items, nil
}
