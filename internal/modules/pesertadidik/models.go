package pesertadidik

type Filters struct {
	SekolahID           string
	RombonganBelajarID  string
	SemesterID          string
	TingkatPendidikanID string
	Search              string
}

type ListItem struct {
	PesertaDidikID string  `json:"peserta_didik_id"`
	Nama           string  `json:"nama"`
	JenisKelamin   string  `json:"jenis_kelamin"`
	NISN           *string `json:"nisn,omitempty"`
	NIPD           *string `json:"nipd,omitempty"`
	SekolahID      string  `json:"sekolah_id"`
	RombelID       *string `json:"rombongan_belajar_id,omitempty"`
	Rombel         *string `json:"rombel,omitempty"`
	JenisRombel    *int64  `json:"jenis_rombel,omitempty"`
	SemesterID     *string `json:"semester_id,omitempty"`
	TingkatID      *int64  `json:"tingkat_pendidikan_id,omitempty"`
	Tingkat        *string `json:"tingkat_pendidikan,omitempty"`
}

type Detail struct {
	PesertaDidikID       string   `json:"peserta_didik_id"`
	Nama                 string   `json:"nama"`
	JenisKelamin         string   `json:"jenis_kelamin"`
	NISN                 *string  `json:"nisn,omitempty"`
	TempatLahir          *string  `json:"tempat_lahir,omitempty"`
	TanggalLahir         *string  `json:"tanggal_lahir,omitempty"`
	AgamaID              *int     `json:"agama_id,omitempty"`
	Agama                *string  `json:"agama,omitempty"`
	AlamatJalan          *string  `json:"alamat_jalan,omitempty"`
	RT                   *int64   `json:"rt,omitempty"`
	RW                   *int64   `json:"rw,omitempty"`
	NamaDusun            *string  `json:"nama_dusun,omitempty"`
	DesaKelurahan        *string  `json:"desa_kelurahan,omitempty"`
	KodeWilayah          *string  `json:"kode_wilayah,omitempty"`
	KodePos              *string  `json:"kode_pos,omitempty"`
	JenisTinggalID       *int64   `json:"jenis_tinggal_id,omitempty"`
	JenisTinggal         *string  `json:"jenis_tinggal,omitempty"`
	AlatTransportasiID   *int64   `json:"alat_transportasi_id,omitempty"`
	AlatTransportasi     *string  `json:"alat_transportasi,omitempty"`
	NomorTeleponSeluler  *string  `json:"nomor_telepon_seluler,omitempty"`
	Email                *string  `json:"email,omitempty"`
	AnakKeberapa         *int64   `json:"anak_keberapa,omitempty"`
	PekerjaanID          *int     `json:"pekerjaan_id,omitempty"`
	Pekerjaan            *string  `json:"pekerjaan,omitempty"`
	SekolahID            *string  `json:"sekolah_id,omitempty"`
	NIPD                 *string  `json:"nipd,omitempty"`
	TanggalMasuk         *string  `json:"tanggal_masuk_sekolah,omitempty"`
	JenisPendaftaranID   *int64   `json:"jenis_pendaftaran_id,omitempty"`
	JenisPendaftaran     *string  `json:"jenis_pendaftaran,omitempty"`
	JenisKeluarID        *string  `json:"jenis_keluar_id,omitempty"`
	JenisKeluar          *string  `json:"jenis_keluar,omitempty"`
	SekolahAsal          *string  `json:"sekolah_asal,omitempty"`
	NamaAyah             *string  `json:"nama_ayah,omitempty"`
	TahunLahirAyah       *int64   `json:"tahun_lahir_ayah,omitempty"`
	PendidikanAyahID     *int64   `json:"jenjang_pendidikan_ayah_id,omitempty"`
	PekerjaanAyahID      *int     `json:"pekerjaan_id_ayah,omitempty"`
	PekerjaanAyah        *string  `json:"pekerjaan_ayah,omitempty"`
	PenghasilanAyahID    *int     `json:"penghasilan_id_ayah,omitempty"`
	PenghasilanAyah      *string  `json:"penghasilan_ayah,omitempty"`
	NamaIbu              *string  `json:"nama_ibu_kandung,omitempty"`
	TahunLahirIbu        *int64   `json:"tahun_lahir_ibu,omitempty"`
	PendidikanIbuID      *int64   `json:"jenjang_pendidikan_ibu_id,omitempty"`
	PekerjaanIbuID       *int     `json:"pekerjaan_id_ibu,omitempty"`
	PekerjaanIbu         *string  `json:"pekerjaan_ibu,omitempty"`
	PenghasilanIbuID     *int     `json:"penghasilan_id_ibu,omitempty"`
	PenghasilanIbu       *string  `json:"penghasilan_ibu,omitempty"`
	NamaWali             *string  `json:"nama_wali,omitempty"`
	TahunLahirWali       *int64   `json:"tahun_lahir_wali,omitempty"`
	PendidikanWaliID     *int64   `json:"jenjang_pendidikan_wali_id,omitempty"`
	PekerjaanWaliID      *int     `json:"pekerjaan_id_wali,omitempty"`
	PekerjaanWali        *string  `json:"pekerjaan_wali,omitempty"`
	PenghasilanWaliID    *int     `json:"penghasilan_id_wali,omitempty"`
	PenghasilanWali      *string  `json:"penghasilan_wali,omitempty"`
	TinggiBadan          *float64 `json:"tinggi_badan,omitempty"`
	BeratBadan           *float64 `json:"berat_badan,omitempty"`
	JarakRumahSekolahKM  *float64 `json:"jarak_rumah_ke_sekolah_km,omitempty"`
	SemesterLongitudinal *string  `json:"semester_longitudinal,omitempty"`
}

type Rombel struct {
	AnggotaRombelID    string  `json:"anggota_rombel_id"`
	RombelID           string  `json:"rombongan_belajar_id"`
	Nama               string  `json:"nama"`
	SemesterID         string  `json:"semester_id"`
	Semester           *string `json:"semester,omitempty"`
	SekolahID          string  `json:"sekolah_id"`
	TingkatID          *int64  `json:"tingkat_pendidikan_id,omitempty"`
	Tingkat            *string `json:"tingkat_pendidikan,omitempty"`
	WaliKelasID        *string `json:"wali_kelas_id,omitempty"`
	WaliKelas          *string `json:"wali_kelas,omitempty"`
	JenisPendaftaranID *int64  `json:"jenis_pendaftaran_id,omitempty"`
	JenisPendaftaran   *string `json:"jenis_pendaftaran,omitempty"`
}

type PrestasiBeasiswa struct {
	Prestasi []Prestasi `json:"prestasi"`
	Beasiswa []Beasiswa `json:"beasiswa"`
}

type Prestasi struct {
	ID                string  `json:"prestasi_id"`
	JenisPrestasiID   int     `json:"jenis_prestasi_id"`
	JenisPrestasi     *string `json:"jenis_prestasi,omitempty"`
	TingkatPrestasiID int     `json:"tingkat_prestasi_id"`
	TingkatPrestasi   *string `json:"tingkat_prestasi,omitempty"`
	Nama              string  `json:"nama"`
	Tahun             int64   `json:"tahun_prestasi"`
	Penyelenggara     *string `json:"penyelenggara,omitempty"`
	Peringkat         *int    `json:"peringkat,omitempty"`
}

type Beasiswa struct {
	ID              string  `json:"beasiswa_peserta_didik_id"`
	JenisBeasiswaID int     `json:"jenis_beasiswa_id"`
	JenisBeasiswa   *string `json:"jenis_beasiswa,omitempty"`
	Keterangan      *string `json:"keterangan,omitempty"`
	TahunMulai      int64   `json:"tahun_mulai"`
	TahunSelesai    *int64  `json:"tahun_selesai,omitempty"`
}
