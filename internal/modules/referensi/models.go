package referensi

type Semester struct {
	SemesterID     string `json:"semester_id"`
	TahunAjaranID  int64  `json:"tahun_ajaran_id"`
	Nama           string `json:"nama"`
	Semester       int    `json:"semester"`
	PeriodeAktif   int    `json:"periode_aktif"`
	TanggalMulai   string `json:"tanggal_mulai"`
	TanggalSelesai string `json:"tanggal_selesai"`
}

type TahunAjaran struct {
	TahunAjaranID  int64  `json:"tahun_ajaran_id"`
	Nama           string `json:"nama"`
	PeriodeAktif   int    `json:"periode_aktif"`
	TanggalMulai   string `json:"tanggal_mulai"`
	TanggalSelesai string `json:"tanggal_selesai"`
}

type Region struct {
	KodeWilayah string `json:"kode_wilayah"`
	Nama        string `json:"nama"`
}

type WilayahHierarchy struct {
	DesaKelurahan *Region `json:"desa_kelurahan,omitempty"`
	Kecamatan     *Region `json:"kecamatan,omitempty"`
	KabupatenKota *Region `json:"kabupaten_kota,omitempty"`
	Provinsi      *Region `json:"provinsi,omitempty"`
}
