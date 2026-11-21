package ptk

type Filters struct {
	SekolahID     string
	TahunAjaranID string
	JenisPTKID    string
	Search        string
}

type ListItem struct {
	PTKID               string  `json:"ptk_id"`
	Nama                string  `json:"nama"`
	NIP                 *string `json:"nip,omitempty"`
	NUPTK               *string `json:"nuptk,omitempty"`
	NRG                 *string `json:"nrg,omitempty"`
	NIK                 string  `json:"nik"`
	NoKK                *string `json:"no_kk,omitempty"`
	NPWP                *string `json:"npwp,omitempty"`
	RekeningBank        *string `json:"rekening_bank,omitempty"`
	RekeningAtasNama    *string `json:"rekening_atas_nama,omitempty"`
	Karpeg              *string `json:"karpeg,omitempty"`
	Karpas              *string `json:"karpas,omitempty"`
	JenisKelamin        string  `json:"jenis_kelamin"`
	StatusKepegawaianID int     `json:"status_kepegawaian_id"`
	StatusKepegawaian   *string `json:"status_kepegawaian,omitempty"`
	SekolahID           string  `json:"sekolah_id"`
	TahunAjaranID       int64   `json:"tahun_ajaran_id"`
	JenisPTKID          int64   `json:"jenis_ptk_id"`
	JenisPTK            *string `json:"jenis_ptk,omitempty"`
}

type Detail struct {
	PTKID               string   `json:"ptk_id"`
	Nama                string   `json:"nama"`
	NIP                 *string  `json:"nip,omitempty"`
	NUPTK               *string  `json:"nuptk,omitempty"`
	NRG                 *string  `json:"nrg,omitempty"`
	NUKS                *string  `json:"nuks,omitempty"`
	NIK                 string   `json:"nik"`
	NoKK                *string  `json:"no_kk,omitempty"`
	NPWP                *string  `json:"npwp,omitempty"`
	RekeningBank        *string  `json:"rekening_bank,omitempty"`
	RekeningAtasNama    *string  `json:"rekening_atas_nama,omitempty"`
	Karpeg              *string  `json:"karpeg,omitempty"`
	Karpas              *string  `json:"karpas,omitempty"`
	JenisKelamin        string   `json:"jenis_kelamin"`
	TempatLahir         string   `json:"tempat_lahir"`
	TanggalLahir        string   `json:"tanggal_lahir"`
	AgamaID             int      `json:"agama_id"`
	Agama               *string  `json:"agama,omitempty"`
	StatusKepegawaianID int      `json:"status_kepegawaian_id"`
	StatusKepegawaian   *string  `json:"status_kepegawaian,omitempty"`
	StatusKeaktifanID   int64    `json:"status_keaktifan_id"`
	StatusKeaktifan     *string  `json:"status_keaktifan,omitempty"`
	PangkatGolonganID   *int64   `json:"pangkat_golongan_id,omitempty"`
	PangkatGolongan     *string  `json:"pangkat_golongan,omitempty"`
	AlamatJalan         string   `json:"alamat_jalan"`
	RT                  *int64   `json:"rt,omitempty"`
	RW                  *int64   `json:"rw,omitempty"`
	NamaDusun           *string  `json:"nama_dusun,omitempty"`
	DesaKelurahan       string   `json:"desa_kelurahan"`
	KodeWilayah         string   `json:"kode_wilayah"`
	NamaWilayah         *string  `json:"nama_wilayah,omitempty"`
	KodePos             *string  `json:"kode_pos,omitempty"`
	Lintang             *float64 `json:"lintang,omitempty"`
	Bujur               *float64 `json:"bujur,omitempty"`
	NoHP                *string  `json:"no_hp,omitempty"`
	Email               *string  `json:"email,omitempty"`
	SekolahID           *string  `json:"sekolah_id,omitempty"`
	TahunAjaranID       *int64   `json:"tahun_ajaran_id,omitempty"`
	JenisPTKID          *int64   `json:"jenis_ptk_id,omitempty"`
	JenisPTK            *string  `json:"jenis_ptk,omitempty"`
	JabatanPTKID        *int64   `json:"jabatan_ptk_id,omitempty"`
	JabatanPTK          *string  `json:"jabatan_ptk,omitempty"`
}

type Pembelajaran struct {
	ID                   string  `json:"pembelajaran_id"`
	SemesterID           string  `json:"semester_id"`
	RombonganBelajarID   string  `json:"rombongan_belajar_id"`
	Rombel               *string `json:"rombel,omitempty"`
	MataPelajaranID      int     `json:"mata_pelajaran_id"`
	MataPelajaran        string  `json:"mata_pelajaran"`
	SKMengajar           string  `json:"sk_mengajar"`
	TanggalSKMengajar    string  `json:"tanggal_sk_mengajar"`
	JamMengajarPerMinggu float64 `json:"jam_mengajar_per_minggu"`
	StatusDiKurikulum    int     `json:"status_di_kurikulum"`
}

type TugasTambahan struct {
	ID           string  `json:"ptk_tugas_tambahan_id"`
	JabatanPTKID int64   `json:"jabatan_ptk_id"`
	Jabatan      *string `json:"jabatan,omitempty"`
	SekolahID    string  `json:"sekolah_id"`
	JumlahJam    float64 `json:"jumlah_jam"`
	NomorSK      string  `json:"nomor_sk"`
	TMT          string  `json:"tmt_tambahan"`
	TST          *string `json:"tst_tambahan,omitempty"`
}

type Riwayat struct {
	PendidikanFormal []PendidikanFormal `json:"pendidikan_formal"`
	Kepangkatan      []Kepangkatan      `json:"kepangkatan"`
	Sertifikasi      []Sertifikasi      `json:"sertifikasi"`
	Diklat           []Diklat           `json:"diklat"`
}

type PendidikanFormal struct {
	ID                  string  `json:"riwayat_pendidikan_formal_id"`
	JenjangPendidikanID int64   `json:"jenjang_pendidikan_id"`
	JenjangPendidikan   *string `json:"jenjang_pendidikan,omitempty"`
	BidangStudiID       int     `json:"bidang_studi_id"`
	BidangStudi         *string `json:"bidang_studi,omitempty"`
	GelarAkademikID     *int    `json:"gelar_akademik_id,omitempty"`
	GelarAkademik       *string `json:"gelar_akademik,omitempty"`
	SatuanPendidikan    string  `json:"satuan_pendidikan_formal"`
	Fakultas            *string `json:"fakultas,omitempty"`
	Kependidikan        int     `json:"kependidikan"`
	TahunMasuk          int     `json:"tahun_masuk"`
	TahunLulus          *int    `json:"tahun_lulus,omitempty"`
	StatusKuliah        int     `json:"status_kuliah"`
	Semester            *int    `json:"semester,omitempty"`
	IPK                 float64 `json:"ipk"`
}

type Kepangkatan struct {
	ID                string  `json:"riwayat_kepangkatan_id"`
	PangkatGolonganID int64   `json:"pangkat_golongan_id"`
	PangkatGolongan   *string `json:"pangkat_golongan,omitempty"`
	NomorSK           string  `json:"nomor_sk"`
	TanggalSK         string  `json:"tanggal_sk"`
	TMTPangkat        string  `json:"tmt_pangkat"`
	MasaKerjaTahun    int     `json:"masa_kerja_gol_tahun"`
	MasaKerjaBulan    int     `json:"masa_kerja_gol_bulan"`
}

type Sertifikasi struct {
	ID                 string  `json:"riwayat_sertifikasi_id"`
	BidangStudiID      int     `json:"bidang_studi_id"`
	BidangStudi        *string `json:"bidang_studi,omitempty"`
	JenisSertifikasiID int64   `json:"jenis_sertifikasi_id"`
	Tanggal            string  `json:"tanggal_sertifikasi"`
	TanggalKedaluwarsa *string `json:"tanggal_kedaluwarsa,omitempty"`
	NomorSertifikat    string  `json:"nomor_sertifikat"`
	NomorRegistrasi    *string `json:"nomor_registrasi,omitempty"`
	NomorPeserta       *string `json:"nomor_peserta,omitempty"`
	Kualifikasi        *string `json:"kualifikasi,omitempty"`
}

type Diklat struct {
	ID            string  `json:"diklat_id"`
	JenisDiklatID int     `json:"jenis_diklat_id"`
	Nama          string  `json:"nama"`
	Penyelenggara *string `json:"penyelenggara,omitempty"`
	Tahun         int     `json:"tahun"`
	Peran         *string `json:"peran,omitempty"`
	Tingkat       *int    `json:"tingkat,omitempty"`
	BerapaJam     int     `json:"berapa_jam"`
	Sertifikat    *string `json:"sertifikat_diklat,omitempty"`
}
