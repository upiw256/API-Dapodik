package rombel

type Filters struct {
	SekolahID           string
	SemesterID          string
	TingkatPendidikanID string
	JenisRombel         string
	KurikulumID         string
	JurusanSPID         string
	PTKID               string
	Search              string
}

type ListItem struct {
	ID                  string  `json:"rombongan_belajar_id"`
	Nama                string  `json:"nama"`
	SemesterID          string  `json:"semester_id"`
	Semester            *string `json:"semester,omitempty"`
	SekolahID           string  `json:"sekolah_id"`
	Sekolah             *string `json:"sekolah,omitempty"`
	TingkatPendidikanID int64   `json:"tingkat_pendidikan_id"`
	TingkatPendidikan   *string `json:"tingkat_pendidikan,omitempty"`
	JenisRombel         int64   `json:"jenis_rombel"`
	JenisRombelNama     *string `json:"jenis_rombel_nama,omitempty"`
	KurikulumID         int64   `json:"kurikulum_id"`
	Kurikulum           *string `json:"kurikulum,omitempty"`
	JurusanSPID         *string `json:"jurusan_sp_id,omitempty"`
	JurusanSP           *string `json:"jurusan_sp,omitempty"`
	Jurusan             *string `json:"jurusan,omitempty"`
	PTKID               *string `json:"ptk_id,omitempty"`
	WaliKelas           *string `json:"wali_kelas,omitempty"`
	RuangID             *string `json:"ruang_id,omitempty"`
	Ruang               *string `json:"ruang,omitempty"`
	JumlahSiswa         int64   `json:"jumlah_siswa_aktif"`
}

type Detail struct {
	ID                      string   `json:"rombongan_belajar_id"`
	Nama                    string   `json:"nama"`
	SemesterID              string   `json:"semester_id"`
	Semester                *string  `json:"semester,omitempty"`
	SekolahID               string   `json:"sekolah_id"`
	Sekolah                 *string  `json:"sekolah,omitempty"`
	TingkatPendidikanID     int64    `json:"tingkat_pendidikan_id"`
	TingkatPendidikan       *string  `json:"tingkat_pendidikan,omitempty"`
	KurikulumID             int64    `json:"kurikulum_id"`
	Kurikulum               *string  `json:"kurikulum,omitempty"`
	JenisRombel             int64    `json:"jenis_rombel"`
	JenisRombelNama         *string  `json:"jenis_rombel_nama,omitempty"`
	MovingClass             *int64   `json:"moving_class,omitempty"`
	SKS                     *float64 `json:"sks,omitempty"`
	TanggalMulai            *string  `json:"tanggal_mulai,omitempty"`
	TanggalSelesai          *string  `json:"tanggal_selesai,omitempty"`
	KebutuhanKhususID       *int     `json:"kebutuhan_khusus_id,omitempty"`
	KebutuhanKhusus         *string  `json:"kebutuhan_khusus,omitempty"`
	PTKID                   *string  `json:"ptk_id,omitempty"`
	WaliKelas               *string  `json:"wali_kelas,omitempty"`
	RuangID                 *string  `json:"ruang_id,omitempty"`
	Ruang                   *string  `json:"ruang,omitempty"`
	JenisPrasarana          *string  `json:"jenis_prasarana,omitempty"`
	BangunanID              *string  `json:"bangunan_id,omitempty"`
	Bangunan                *string  `json:"bangunan,omitempty"`
	JurusanSPID             *string  `json:"jurusan_sp_id,omitempty"`
	JurusanSP               *string  `json:"jurusan_sp,omitempty"`
	Jurusan                 *string  `json:"jurusan,omitempty"`
	JumlahSiswa             int64    `json:"jumlah_siswa_aktif"`
	StatusValidasi          *int64   `json:"status_validasi,omitempty"`
	JumlahValidasi          int64    `json:"jumlah_validasi"`
	EkstrakurikulerTersedia bool     `json:"ekstrakurikuler_tersedia"`
}

type Anggota struct {
	AnggotaRombelID          string  `json:"anggota_rombel_id"`
	PesertaDidikID           string  `json:"peserta_didik_id"`
	Nama                     string  `json:"nama"`
	JenisKelamin             string  `json:"jenis_kelamin"`
	NISN                     *string `json:"nisn,omitempty"`
	NIPD                     *string `json:"nipd,omitempty"`
	TanggalMasuk             *string `json:"tanggal_masuk_sekolah,omitempty"`
	JenisPendaftaranRombelID *int64  `json:"jenis_pendaftaran_rombel_id,omitempty"`
	JenisPendaftaranRombel   *string `json:"jenis_pendaftaran_rombel,omitempty"`
	JenisPendaftaranAwalID   *int64  `json:"jenis_pendaftaran_awal_id,omitempty"`
	JenisPendaftaranAwal     *string `json:"jenis_pendaftaran_awal,omitempty"`
	AgamaID                  *int    `json:"agama_id,omitempty"`
	Agama                    *string `json:"agama,omitempty"`
	JenisKeluarID            *string `json:"jenis_keluar_id,omitempty"`
	JenisKeluar              *string `json:"jenis_keluar,omitempty"`
}

type AnggotaResponse struct {
	RombelID  string    `json:"rombongan_belajar_id"`
	NamaKelas string    `json:"nama_kelas"`
	WaliKelas *string   `json:"wali_kelas,omitempty"`
	Anggota   []Anggota `json:"anggota"`
}

type Pembelajaran struct {
	ID                  string  `json:"pembelajaran_id"`
	RombelID            string  `json:"rombongan_belajar_id,omitempty"`
	NamaKelas           string  `json:"nama_kelas,omitempty"`
	WaliKelas           *string `json:"wali_kelas,omitempty"`
	SemesterID          string  `json:"semester_id"`
	MataPelajaranID     int     `json:"mata_pelajaran_id"`
	KodeMataPelajaran   *string `json:"kode_mata_pelajaran,omitempty"`
	MataPelajaran       string  `json:"mata_pelajaran"`
	PTKID               string  `json:"ptk_id"`
	NamaGuru            string  `json:"nama_guru"`
	PTKTerdaftarID      string  `json:"ptk_terdaftar_id"`
	SKMengajar          string  `json:"sk_mengajar"`
	TanggalSK           *string `json:"tanggal_sk_mengajar,omitempty"`
	JamPerMinggu        float64 `json:"jam_mengajar_per_minggu"`
	StatusDiKurikulum   int     `json:"status_di_kurikulum"`
	Induk               bool    `json:"induk"`
	IndukPembelajaranID *string `json:"induk_pembelajaran_id,omitempty"`
}

type Distribution struct {
	ID     int    `json:"id"`
	Label  string `json:"label"`
	Jumlah int64  `json:"jumlah"`
}
type AgeDistribution struct {
	Rentang string `json:"rentang"`
	Jumlah  int64  `json:"jumlah"`
}
type Statistics struct {
	TotalSiswa            int64             `json:"total_siswa"`
	LakiLaki              int64             `json:"laki_laki"`
	Perempuan             int64             `json:"perempuan"`
	TidakDiketahui        int64             `json:"jenis_kelamin_tidak_diketahui"`
	Agama                 []Distribution    `json:"sebaran_agama"`
	Usia                  []AgeDistribution `json:"sebaran_usia"`
	JumlahPembelajaran    int64             `json:"jumlah_pembelajaran"`
	JumlahMataPelajaran   int64             `json:"jumlah_mata_pelajaran"`
	EvaluasiRaporTersedia bool              `json:"evaluasi_rapor_tersedia"`
}
