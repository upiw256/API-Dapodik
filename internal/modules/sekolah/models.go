package sekolah

type Profile struct {
	SekolahID           string         `json:"sekolah_id"`
	NPSN                *string        `json:"npsn,omitempty"`
	Nama                string         `json:"nama"`
	NamaNomenklatur     *string        `json:"nama_nomenklatur,omitempty"`
	BentukPendidikanID  int            `json:"bentuk_pendidikan_id"`
	BentukPendidikan    *string        `json:"bentuk_pendidikan,omitempty"`
	StatusSekolah       int            `json:"status_sekolah"`
	StatusKepemilikanID int64          `json:"status_kepemilikan_id"`
	StatusKepemilikan   *string        `json:"status_kepemilikan,omitempty"`
	AlamatJalan         string         `json:"alamat_jalan"`
	RT                  *int64         `json:"rt,omitempty"`
	RW                  *int64         `json:"rw,omitempty"`
	NamaDusun           *string        `json:"nama_dusun,omitempty"`
	DesaKelurahan       string         `json:"desa_kelurahan"`
	KodeWilayah         string         `json:"kode_wilayah"`
	NamaWilayah         *string        `json:"nama_wilayah,omitempty"`
	KodePos             *string        `json:"kode_pos,omitempty"`
	Lintang             *float64       `json:"lintang,omitempty"`
	Bujur               *float64       `json:"bujur,omitempty"`
	NomorTelepon        *string        `json:"nomor_telepon,omitempty"`
	NomorFax            *string        `json:"nomor_fax,omitempty"`
	Email               *string        `json:"email,omitempty"`
	Website             *string        `json:"website,omitempty"`
	LuasTanahMilik      float64        `json:"luas_tanah_milik"`
	LuasTanahBukanMilik float64        `json:"luas_tanah_bukan_milik"`
	Akreditasi          *Accreditation `json:"akreditasi,omitempty"`
}

type Accreditation struct {
	ID            int64   `json:"akreditasi_id"`
	Nama          *string `json:"nama,omitempty"`
	SK            string  `json:"sk"`
	Tanggal       string  `json:"tanggal"`
	BerlakuSampai string  `json:"berlaku_sampai"`
}

type Longitudinal struct {
	SekolahID              string     `json:"sekolah_id"`
	SemesterID             string     `json:"semester_id"`
	WaktuPenyelenggaraanID int64      `json:"waktu_penyelenggaraan_id"`
	WaktuPenyelenggaraan   *string    `json:"waktu_penyelenggaraan,omitempty"`
	KontinuitasListrik     string     `json:"kontinuitas_listrik"`
	JarakListrik           *float64   `json:"jarak_listrik,omitempty"`
	WilayahTerpencil       int        `json:"wilayah_terpencil"`
	WilayahPerbatasan      int        `json:"wilayah_perbatasan"`
	WilayahTransmigrasi    int        `json:"wilayah_transmigrasi"`
	WilayahAdatTerpencil   int        `json:"wilayah_adat_terpencil"`
	WilayahBencanaAlam     int        `json:"wilayah_bencana_alam"`
	WilayahBencanaSosial   int        `json:"wilayah_bencana_sosial"`
	PartisipasiBOS         int        `json:"partisipasi_bos"`
	SertifikasiISOID       int        `json:"sertifikasi_iso_id"`
	SumberListrikID        int64      `json:"sumber_listrik_id"`
	SumberListrik          *string    `json:"sumber_listrik,omitempty"`
	DayaListrik            int64      `json:"daya_listrik"`
	AksesInternetID        *int       `json:"akses_internet_id,omitempty"`
	AksesInternet          *string    `json:"akses_internet,omitempty"`
	AksesInternet2ID       int        `json:"akses_internet_2_id"`
	AksesInternet2         *string    `json:"akses_internet_2,omitempty"`
	Listrik                []Listrik  `json:"listrik"`
	Internet               []Internet `json:"internet"`
	Sanitasi               *Sanitasi  `json:"sanitasi,omitempty"`
}

type Listrik struct {
	ID              string  `json:"listrik_id"`
	SumberListrikID int64   `json:"sumber_listrik_id"`
	SumberListrik   *string `json:"sumber_listrik,omitempty"`
	Daya            int64   `json:"daya"`
	Kontinuitas     string  `json:"kontinuitas_listrik"`
	StatusSambungan string  `json:"status_sambungan"`
	Utama           int     `json:"utama"`
}

type Internet struct {
	ID             string   `json:"internet_id"`
	Provider       string   `json:"provider"`
	JenisLayananID *int64   `json:"jenis_layanan_internet_id,omitempty"`
	JenisKoneksiID *int64   `json:"jenis_koneksi_id,omitempty"`
	Bandwidth      float64  `json:"bandwidth"`
	BandwidthUp    *float64 `json:"bandwidth_up,omitempty"`
	BandwidthDown  *float64 `json:"bandwidth_down,omitempty"`
	Latency        *int     `json:"latency,omitempty"`
}

type Sanitasi struct {
	SemesterID            string `json:"semester_id"`
	SumberAirID           int64  `json:"sumber_air_id"`
	SumberAirMinumID      int64  `json:"sumber_air_minum_id"`
	KetersediaanAir       int    `json:"ketersediaan_air"`
	KecukupanAir          int    `json:"kecukupan_air"`
	MinumSiswa            int    `json:"minum_siswa"`
	ToiletSiswaLaki       int    `json:"toilet_siswa_laki"`
	ToiletSiswaPerempuan  int    `json:"toilet_siswa_perempuan"`
	ToiletSiswaKhusus     int    `json:"toilet_siswa_kk"`
	ToiletSiswaKecil      int    `json:"toilet_siswa_kecil"`
	TempatCuciTangan      int    `json:"tempat_cuci_tangan"`
	AksesSabunAirMengalir int    `json:"a_sabun_air_mengalir"`
	TipeJamban            string `json:"tipe_jamban"`
}

type Jurusan struct {
	JurusanSPID   string  `json:"jurusan_sp_id"`
	JurusanID     string  `json:"jurusan_id"`
	Nama          string  `json:"nama_jurusan_sp"`
	NamaReferensi *string `json:"nama_jurusan,omitempty"`
	SKIzin        *string `json:"sk_izin,omitempty"`
	TanggalSK     *string `json:"tanggal_sk_izin,omitempty"`
}

type Rombel struct {
	ID                  string  `json:"rombongan_belajar_id"`
	SemesterID          string  `json:"semester_id"`
	Nama                string  `json:"nama"`
	TingkatPendidikanID int64   `json:"tingkat_pendidikan_id"`
	TingkatPendidikan   *string `json:"tingkat_pendidikan,omitempty"`
	KurikulumID         int     `json:"kurikulum_id"`
	Kurikulum           *string `json:"kurikulum,omitempty"`
	RuangID             *string `json:"ruang_id,omitempty"`
	Ruang               *string `json:"ruang,omitempty"`
	PTKID               *string `json:"ptk_id,omitempty"`
	WaliKelas           *string `json:"wali_kelas,omitempty"`
	TanggalMulai        string  `json:"tanggal_mulai"`
	TanggalSelesai      string  `json:"tanggal_selesai"`
	JenisRombel         int     `json:"jenis_rombel"`
}

type Room struct {
	ID               string   `json:"ruang_id"`
	JenisPrasaranaID int      `json:"jenis_prasarana_id"`
	JenisPrasarana   *string  `json:"jenis_prasarana,omitempty"`
	Kode             string   `json:"kode_ruang"`
	Nama             string   `json:"nama_ruang"`
	Lantai           int      `json:"lantai"`
	Panjang          *float64 `json:"panjang,omitempty"`
	Lebar            *float64 `json:"lebar,omitempty"`
	Kapasitas        *int64   `json:"kapasitas,omitempty"`
	Luas             *float64 `json:"luas_ruang,omitempty"`
	BangunanID       *string  `json:"bangunan_id,omitempty"`
	Bangunan         *string  `json:"bangunan,omitempty"`
	JumlahLantai     *int64   `json:"jumlah_lantai_bangunan,omitempty"`
}

type SchoolList struct {
	Items []Profile `json:"items"`
	Page  int       `json:"page"`
	Limit int       `json:"limit"`
	Total int64     `json:"total"`
}
