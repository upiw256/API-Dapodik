package sekolah

import (
	"context"
	"errors"
	"fmt"

	"dapodik-readonly-api/internal/period"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("sekolah tidak ditemukan")

type Store interface {
	ListSchools(context.Context, int, int, string) ([]Profile, int64, error)
	GetSchool(context.Context, string) (Profile, error)
	GetLongitudinal(context.Context, string, string) ([]Longitudinal, error)
	ListJurusan(context.Context, string) ([]Jurusan, error)
	ListRombel(context.Context, string, string) ([]Rombel, error)
	ListRooms(context.Context, string) ([]Room, error)
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

const profileSelect = `
	SELECT s.sekolah_id::text,
	       NULLIF(btrim(s.npsn), ''),
	       s.nama,
	       s.nama_nomenklatur,
	       s.bentuk_pendidikan_id::integer,
	       bp.nama,
	       s.status_sekolah::integer,
	       s.status_kepemilikan_id::bigint,
	       sk.nama,
	       s.alamat_jalan,
	       s.rt::bigint,
	       s.rw::bigint,
	       s.nama_dusun,
	       s.desa_kelurahan,
	       btrim(s.kode_wilayah),
	       mw.nama,
	       NULLIF(btrim(s.kode_pos), ''),
	       s.lintang::double precision,
	       s.bujur::double precision,
	       s.nomor_telepon,
	       s.nomor_fax,
	       s.email,
	       s.website,
	       s.luas_tanah_milik::double precision,
	       s.luas_tanah_bukan_milik::double precision,
	       ak.akreditasi_id,
	       ak.nama_akreditasi,
	       ak.sk,
	       ak.tanggal,
	       ak.berlaku_sampai
	FROM public.sekolah AS s
	LEFT JOIN ref.bentuk_pendidikan AS bp
	  ON bp.bentuk_pendidikan_id = s.bentuk_pendidikan_id
	 AND bp.expired_date IS NULL
	LEFT JOIN ref.status_kepemilikan AS sk
	  ON sk.status_kepemilikan_id = s.status_kepemilikan_id
	 AND sk.expired_date IS NULL
	LEFT JOIN ref.mst_wilayah AS mw
	  ON mw.kode_wilayah = s.kode_wilayah
	 AND mw.expired_date IS NULL
	LEFT JOIN (
		SELECT DISTINCT ON (a.sekolah_id)
		       a.sekolah_id, a.akreditasi_id::bigint,
		       ar.nama AS nama_akreditasi, a.akred_sp_sk AS sk,
		       to_char(a.akred_sp_tmt, 'YYYY-MM-DD') AS tanggal,
		       to_char(a.akred_sp_tst, 'YYYY-MM-DD') AS berlaku_sampai
		FROM public.akreditasi_sp AS a
		LEFT JOIN ref.akreditasi AS ar
		  ON ar.akreditasi_id = a.akreditasi_id
		 AND ar.expired_date IS NULL
		WHERE a.soft_delete = 0
		ORDER BY a.sekolah_id, a.akred_sp_tmt DESC, a.last_update DESC
	) AS ak ON ak.sekolah_id = s.sekolah_id
	WHERE s.soft_delete = 0`

func (r *PostgresRepository) ListSchools(ctx context.Context, page, limit int, npsn string) ([]Profile, int64, error) {
	var total int64
	if err := r.pool.QueryRow(ctx, `
		SELECT count(*)
		FROM public.sekolah
		WHERE soft_delete = 0
		  AND ($1::text = '' OR btrim(npsn) = $1::text)`, npsn).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count schools: %w", err)
	}

	query := profileSelect + `
	  AND ($1::text = '' OR btrim(s.npsn) = $1::text)
	ORDER BY s.nama, s.sekolah_id`
	args := []any{npsn}
	if limit > 0 {
		query += ` LIMIT $2 OFFSET $3`
		args = append(args, limit, (page-1)*limit)
	}
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query schools: %w", err)
	}
	defer rows.Close()

	items := make([]Profile, 0)
	for rows.Next() {
		item, err := scanProfile(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan school: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate schools: %w", err)
	}
	return items, total, nil
}

func (r *PostgresRepository) GetSchool(ctx context.Context, id string) (Profile, error) {
	query := profileSelect + ` AND s.sekolah_id = $1::uuid`
	item, err := scanProfile(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Profile{}, ErrNotFound
		}
		return Profile{}, fmt.Errorf("query school: %w", err)
	}
	return item, nil
}

func scanProfile(row pgx.Row) (Profile, error) {
	var item Profile
	var accreditationID *int64
	var accreditationName, accreditationSK, accreditationDate, accreditationExpiry *string
	err := row.Scan(
		&item.SekolahID,
		&item.NPSN,
		&item.Nama,
		&item.NamaNomenklatur,
		&item.BentukPendidikanID,
		&item.BentukPendidikan,
		&item.StatusSekolah,
		&item.StatusKepemilikanID,
		&item.StatusKepemilikan,
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
		&item.NomorTelepon,
		&item.NomorFax,
		&item.Email,
		&item.Website,
		&item.LuasTanahMilik,
		&item.LuasTanahBukanMilik,
		&accreditationID,
		&accreditationName,
		&accreditationSK,
		&accreditationDate,
		&accreditationExpiry,
	)
	if err != nil {
		return Profile{}, err
	}
	if accreditationID != nil {
		item.Akreditasi = &Accreditation{
			ID:            *accreditationID,
			Nama:          accreditationName,
			SK:            valueOrEmpty(accreditationSK),
			Tanggal:       valueOrEmpty(accreditationDate),
			BerlakuSampai: valueOrEmpty(accreditationExpiry),
		}
	}
	return item, nil
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func (r *PostgresRepository) GetLongitudinal(ctx context.Context, id, semesterID string) ([]Longitudinal, error) {
	if r.period.SemesterID != "" {
		semesterID = r.period.SemesterID
	}
	const query = `
		SELECT sl.sekolah_id::text,
		       btrim(sl.semester_id),
		       sl.waktu_penyelenggaraan_id::bigint,
	       wp.nama,
	       btrim(sl.kontinuitas_listrik),
	       sl.jarak_listrik::double precision,
	       sl.wilayah_terpencil::integer,
	       sl.wilayah_perbatasan::integer,
	       sl.wilayah_transmigrasi::integer,
	       sl.wilayah_adat_terpencil::integer,
	       sl.wilayah_bencana_alam::integer,
	       sl.wilayah_bencana_sosial::integer,
	       sl.partisipasi_bos::integer,
	       sl.sertifikasi_iso_id::integer,
	       sl.sumber_listrik_id::bigint,
	       sumber.nama,
	       sl.daya_listrik::bigint,
	       sl.akses_internet_id::integer,
	       internet.nama,
	       sl.akses_internet_2_id::integer,
	       internet_kedua.nama
	FROM public.sekolah_longitudinal AS sl
	LEFT JOIN ref.waktu_penyelenggaraan AS wp
	  ON wp.waktu_penyelenggaraan_id = sl.waktu_penyelenggaraan_id
	 AND wp.expired_date IS NULL
	LEFT JOIN ref.sumber_listrik AS sumber
	  ON sumber.sumber_listrik_id = sl.sumber_listrik_id
	 AND sumber.expired_date IS NULL
	LEFT JOIN ref.akses_internet AS internet
	  ON internet.akses_internet_id = sl.akses_internet_id
	 AND internet.expired_date IS NULL
	LEFT JOIN ref.akses_internet AS internet_kedua
	  ON internet_kedua.akses_internet_id = sl.akses_internet_2_id
	 AND internet_kedua.expired_date IS NULL
	WHERE sl.sekolah_id = $1::uuid
	  AND sl.soft_delete = 0
	  AND ($2::text = '' OR btrim(sl.semester_id) = $2::text)
	ORDER BY sl.semester_id DESC`

	rows, err := r.pool.Query(ctx, query, id, semesterID)
	if err != nil {
		return nil, fmt.Errorf("query school longitudinal: %w", err)
	}
	defer rows.Close()

	items := make([]Longitudinal, 0)
	for rows.Next() {
		var item Longitudinal
		if err := rows.Scan(
			&item.SekolahID,
			&item.SemesterID,
			&item.WaktuPenyelenggaraanID,
			&item.WaktuPenyelenggaraan,
			&item.KontinuitasListrik,
			&item.JarakListrik,
			&item.WilayahTerpencil,
			&item.WilayahPerbatasan,
			&item.WilayahTransmigrasi,
			&item.WilayahAdatTerpencil,
			&item.WilayahBencanaAlam,
			&item.WilayahBencanaSosial,
			&item.PartisipasiBOS,
			&item.SertifikasiISOID,
			&item.SumberListrikID,
			&item.SumberListrik,
			&item.DayaListrik,
			&item.AksesInternetID,
			&item.AksesInternet,
			&item.AksesInternet2ID,
			&item.AksesInternet2,
		); err != nil {
			return nil, fmt.Errorf("scan school longitudinal: %w", err)
		}
		item.Listrik = make([]Listrik, 0)
		item.Internet = make([]Internet, 0)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate school longitudinal: %w", err)
	}
	if len(items) == 0 {
		return items, nil
	}

	electricity, err := r.listElectricity(ctx, id)
	if err != nil {
		return nil, err
	}
	internet, err := r.listInternet(ctx, id)
	if err != nil {
		return nil, err
	}
	sanitation, err := r.listSanitation(ctx, id, semesterID)
	if err != nil {
		return nil, err
	}
	for i := range items {
		items[i].Listrik = electricity
		items[i].Internet = internet
		items[i].Sanitasi = sanitation[items[i].SemesterID]
	}
	return items, nil
}

func (r *PostgresRepository) listElectricity(ctx context.Context, id string) ([]Listrik, error) {
	const query = `
		SELECT l.listrik_id::text,
		       l.sumber_listrik_id::bigint,
		       sl.nama,
		       l.daya::bigint,
		       btrim(l.kontinuitas_listrik),
		       l.status_sambungan,
		       l.utama::integer
		FROM public.listrik AS l
		LEFT JOIN ref.sumber_listrik AS sl
		  ON sl.sumber_listrik_id = l.sumber_listrik_id
		 AND sl.expired_date IS NULL
		WHERE l.sekolah_id = $1::uuid
		  AND l.soft_delete = 0
		ORDER BY l.utama DESC, l.create_date DESC`

	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("query school electricity: %w", err)
	}
	defer rows.Close()

	items := make([]Listrik, 0)
	for rows.Next() {
		var item Listrik
		if err := rows.Scan(&item.ID, &item.SumberListrikID, &item.SumberListrik, &item.Daya, &item.Kontinuitas, &item.StatusSambungan, &item.Utama); err != nil {
			return nil, fmt.Errorf("scan school electricity: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate school electricity: %w", err)
	}
	return items, nil
}

func (r *PostgresRepository) listInternet(ctx context.Context, id string) ([]Internet, error) {
	const query = `
		SELECT internet_id::text,
		       provider,
		       jenis_layanan_internet_id::bigint,
		       jenis_koneksi_id::bigint,
		       bandwidth::double precision,
		       bandwidth_up::double precision,
		       bandwidth_down::double precision,
		       latency::integer
		FROM public.internet
		WHERE sekolah_id = $1::uuid
		  AND soft_delete = 0
		ORDER BY create_date DESC`

	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("query school internet: %w", err)
	}
	defer rows.Close()

	items := make([]Internet, 0)
	for rows.Next() {
		var item Internet
		if err := rows.Scan(&item.ID, &item.Provider, &item.JenisLayananID, &item.JenisKoneksiID, &item.Bandwidth, &item.BandwidthUp, &item.BandwidthDown, &item.Latency); err != nil {
			return nil, fmt.Errorf("scan school internet: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate school internet: %w", err)
	}
	return items, nil
}

func (r *PostgresRepository) listSanitation(ctx context.Context, id, semesterID string) (map[string]*Sanitasi, error) {
	const query = `
		SELECT btrim(semester_id),
		       sumber_air_id::bigint,
		       sumber_air_minum_id::bigint,
		       ketersediaan_air::integer,
		       kecukupan_air::integer,
		       minum_siswa::integer,
		       toilet_siswa_laki::integer,
		       toilet_siswa_perempuan::integer,
		       toilet_siswa_kk::integer,
		       toilet_siswa_kecil::integer,
		       tempat_cuci_tangan::integer,
		       a_sabun_air_mengalir::integer,
		       btrim(tipe_jamban)
		FROM public.sanitasi
		WHERE sekolah_id = $1::uuid
		  AND soft_delete = 0
		  AND ($2::text = '' OR btrim(semester_id) = $2::text)`

	rows, err := r.pool.Query(ctx, query, id, semesterID)
	if err != nil {
		return nil, fmt.Errorf("query school sanitation: %w", err)
	}
	defer rows.Close()

	items := make(map[string]*Sanitasi)
	for rows.Next() {
		var item Sanitasi
		if err := rows.Scan(
			&item.SemesterID,
			&item.SumberAirID,
			&item.SumberAirMinumID,
			&item.KetersediaanAir,
			&item.KecukupanAir,
			&item.MinumSiswa,
			&item.ToiletSiswaLaki,
			&item.ToiletSiswaPerempuan,
			&item.ToiletSiswaKhusus,
			&item.ToiletSiswaKecil,
			&item.TempatCuciTangan,
			&item.AksesSabunAirMengalir,
			&item.TipeJamban,
		); err != nil {
			return nil, fmt.Errorf("scan school sanitation: %w", err)
		}
		items[item.SemesterID] = &item
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate school sanitation: %w", err)
	}
	return items, nil
}

func (r *PostgresRepository) ListJurusan(ctx context.Context, id string) ([]Jurusan, error) {
	const query = `
		SELECT js.jurusan_sp_id::text,
		       js.jurusan_id,
		       js.nama_jurusan_sp,
		       j.nama_jurusan,
		       js.sk_izin,
		       to_char(js.tanggal_sk_izin, 'YYYY-MM-DD')
		FROM public.jurusan_sp AS js
		LEFT JOIN ref.jurusan AS j
		  ON j.jurusan_id = js.jurusan_id
		 AND j.expired_date IS NULL
		WHERE js.sekolah_id = $1::uuid
		  AND js.soft_delete = 0
		ORDER BY js.nama_jurusan_sp, js.jurusan_sp_id`

	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("query school majors: %w", err)
	}
	defer rows.Close()

	items := make([]Jurusan, 0)
	for rows.Next() {
		var item Jurusan
		if err := rows.Scan(&item.JurusanSPID, &item.JurusanID, &item.Nama, &item.NamaReferensi, &item.SKIzin, &item.TanggalSK); err != nil {
			return nil, fmt.Errorf("scan school major: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate school majors: %w", err)
	}
	return items, nil
}

func (r *PostgresRepository) ListRombel(ctx context.Context, id, semesterID string) ([]Rombel, error) {
	if r.period.SemesterID != "" {
		semesterID = r.period.SemesterID
	}
	const query = `
		SELECT rb.rombongan_belajar_id::text,
		       btrim(rb.semester_id),
		       rb.nama,
		       rb.tingkat_pendidikan_id::bigint,
		       tp.nama,
		       rb.kurikulum_id::integer,
		       k.nama_kurikulum,
		       ruang.id_ruang::text,
		       ruang.nm_ruang,
		       p.ptk_id::text,
		       p.nama,
		       to_char(rb.tanggal_mulai, 'YYYY-MM-DD'),
		       to_char(rb.tanggal_selesai, 'YYYY-MM-DD'),
		       rb.jenis_rombel::integer
		FROM public.rombongan_belajar AS rb
		LEFT JOIN ref.tingkat_pendidikan AS tp
		  ON tp.tingkat_pendidikan_id = rb.tingkat_pendidikan_id
		 AND tp.expired_date IS NULL
		LEFT JOIN ref.kurikulum AS k
		  ON k.kurikulum_id = rb.kurikulum_id
		 AND k.expired_date IS NULL
		LEFT JOIN public.ruang AS ruang
		  ON ruang.id_ruang = rb.id_ruang
		 AND ruang.soft_delete = 0
		LEFT JOIN public.ptk AS p
		  ON p.ptk_id = rb.ptk_id
		 AND p.soft_delete = 0
		WHERE rb.sekolah_id = $1::uuid
		  AND rb.soft_delete = 0
		  AND ($2::text = '' OR btrim(rb.semester_id) = $2::text)
		ORDER BY rb.semester_id DESC, rb.nama`

	rows, err := r.pool.Query(ctx, query, id, semesterID)
	if err != nil {
		return nil, fmt.Errorf("query school rombel: %w", err)
	}
	defer rows.Close()

	items := make([]Rombel, 0)
	for rows.Next() {
		var item Rombel
		if err := rows.Scan(
			&item.ID,
			&item.SemesterID,
			&item.Nama,
			&item.TingkatPendidikanID,
			&item.TingkatPendidikan,
			&item.KurikulumID,
			&item.Kurikulum,
			&item.RuangID,
			&item.Ruang,
			&item.PTKID,
			&item.WaliKelas,
			&item.TanggalMulai,
			&item.TanggalSelesai,
			&item.JenisRombel,
		); err != nil {
			return nil, fmt.Errorf("scan school rombel: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate school rombel: %w", err)
	}
	return items, nil
}

func (r *PostgresRepository) ListRooms(ctx context.Context, id string) ([]Room, error) {
	const query = `
		SELECT ruang.id_ruang::text,
		       ruang.jenis_prasarana_id,
		       jp.nama,
		       ruang.kd_ruang,
		       ruang.nm_ruang,
		       ruang.lantai::integer,
		       ruang.panjang,
		       ruang.lebar,
		       ruang.kapasitas::bigint,
		       ruang.luas_ruang,
		       bangunan.id_bangunan::text,
		       bangunan.nama,
		       bangunan.jml_lantai::bigint
		FROM public.ruang AS ruang
		LEFT JOIN public.bangunan AS bangunan
		  ON bangunan.id_bangunan = ruang.id_bangunan
		 AND bangunan.soft_delete = 0
		LEFT JOIN ref.jenis_prasarana AS jp
		  ON jp.jenis_prasarana_id = ruang.jenis_prasarana_id
		 AND jp.expired_date IS NULL
		WHERE ruang.sekolah_id = $1::uuid
		  AND ruang.soft_delete = 0
		ORDER BY bangunan.nama, ruang.lantai, ruang.nm_ruang`

	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("query school rooms: %w", err)
	}
	defer rows.Close()

	items := make([]Room, 0)
	for rows.Next() {
		var item Room
		if err := rows.Scan(
			&item.ID,
			&item.JenisPrasaranaID,
			&item.JenisPrasarana,
			&item.Kode,
			&item.Nama,
			&item.Lantai,
			&item.Panjang,
			&item.Lebar,
			&item.Kapasitas,
			&item.Luas,
			&item.BangunanID,
			&item.Bangunan,
			&item.JumlahLantai,
		); err != nil {
			return nil, fmt.Errorf("scan school room: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate school rooms: %w", err)
	}
	return items, nil
}
