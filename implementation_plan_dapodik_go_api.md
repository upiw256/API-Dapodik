# 📋 Implementation Plan: Read-Only Dapodik REST API (Go / Golang)

## 1. Instruksi Eksekusi & Tech Stack untuk AI Agent (Gravity)
> **SYSTEM PROMPT / ATURAN EKSEKUSI UNTUK AGENT:**
> 1. **Tech Stack Wajib:**
>    - **Language:** Go (Golang 1.22+)
>    - **HTTP Router/Framework:** `chi` (`github.com/go-chi/chi/v5`) atau `stdlib net/http` agar idiomatik, ringan, dan mudah di-maintain.
>    - **Database Driver:** `pgx/v5` (`github.com/jackc/pgx/v5/pgxpool`) dengan *parameterized raw SQL queries* (`$1, $2, ...`) — **DILARANG** menggunakan ORM berat atau auto-migration.
>    - **Config & Env:** `godotenv` / `envconfig`.
> 2. **Eksekusi Bertahap:** Kerjakan tugas ini **secara berurutan (Step-by-Step)** dari **Fase 1 hingga Fase 5**. **JANGAN melompat ke fase berikutnya** sebelum seluruh item ceklis pada fase saat ini selesai, terkompilasi bersih (`go build ./...` & `go vet ./...`), dan ditandai selesai (`[x]`).
> 3. **STRICT READ-ONLY POLICY:** Backend ini **HANYA** melayani operasi baca (`GET`). Dilarang keras membuat *handler* atau *route* untuk mutasi (`POST`, `PUT`, `PATCH`, `DELETE`) atau menjalankan perintah DDL/DML pada database `pendataan`.
> 4. **BATASAN DOMAIN DATA:** Fokus eksklusif pada 3 entitas utama beserta relasi dan tabel referensinya (`ref.*`):
>    - **Sekolah** (`public.sekolah`)
>    - **Guru / PTK** (`public.ptk`)
>    - **Siswa / Peserta Didik** (`public.peserta_didik`)
> 5. **DATA FILTERING & SANITIZATION (STRUCT & SQL LEVEL):**
>    - Selalu terapkan filter `soft_delete = 0` pada setiap *query* tabel di schema `public` (dan `expired_date IS NULL` pada tabel schema `ref` jika relevan).
>    - **Jangan mendefinisikan field sensitif pada Go Response Struct atau klausa `SELECT` SQL**, kecuali kolom PTK yang secara eksplisit diizinkan pada Fase 4: `nik`, `no_kk`, `npwp`, `rekening_bank`, `rekening_atas_nama`, `karpeg`, dan `karpas`. Kolom sensitif lainnya tetap diblokir: `no_rekening`, `nik_ayah`, `nik_ibu`, `nik_wali`, `no_kps`, `no_kip`, `no_kks`, dan `blob_id`.

---

## 2. Arsitektur Proyek & Pemetaan Skema Database (`pendataan`)

### Struktur Direktori Standar Go
```text
dapodik-readonly-api/
├── cmd/
│   └── api/
│       └── main.go                 # Entrypoint, graceful shutdown, router setup
├── internal/
│   ├── config/
│   │   └── config.go               # Load .env (DB_DSN, APP_PORT, API_KEY)
│   ├── database/
│   │   └── postgres.go             # pgxpool setup + default_transaction_read_only=on
│   ├── middleware/
│   │   ├── readonly.go             # Tolak HTTP method selain GET/HEAD/OPTIONS (405)
│   │   └── auth.go                 # Validasi X-API-Key / Bearer Token
│   ├── response/
│   │   └── json.go                 # Standard JSON wrapper & error helper
│   ├── modules/
│   │   ├── referensi/              # Handler, Service, Repository untuk ref.*
│   │   ├── sekolah/                # Handler, Service, Repository untuk public.sekolah
│   │   ├── ptk/                    # Handler, Service, Repository untuk public.ptk
│   │   └── pesertadidik/           # Handler, Service, Repository untuk public.peserta_didik
│   └── utils/
│       └── pagination.go           # Helper parsing limit, page, offset, dan search
├── .env.example
├── go.mod
├── go.sum
└── API_DOCS.md
```

### Pemetaan Domain & Tabel Referensi (`pendataan`)
1. **Domain Sekolah (`public.sekolah`):**
   - **Utama:** `public.sekolah` (`sekolah_id`), `public.sekolah_longitudinal` (`semester_id`), `public.akreditasi_sp`, `public.izin_operasional`, `public.jurusan_sp`, `public.rombongan_belajar`, `public.bangunan`, `public.ruang`, `public.tanah`, `public.listrik`, `public.internet`, `public.sanitasi`.
   - **Referensi (`ref.*`):** `bentuk_pendidikan`, `status_kepemilikan`, `mst_wilayah`, `akreditasi`, `waktu_penyelenggaraan`, `sumber_listrik`, `akses_internet`, `kurikulum`, `jurusan`, `jenis_prasarana`.
2. **Domain Guru / PTK (`public.ptk`):**
   - **Utama:** `public.ptk` (`ptk_id`), `public.ptk_terdaftar` (`tahun_ajaran_id`), `public.pembelajaran` (`semester_id`), `public.tugas_tambahan`, `public.rwy_pend_formal`, `public.rwy_kepangkatan`, `public.rwy_fungsional`, `public.rwy_struktural`, `public.rwy_kerja`, `public.rwy_sertifikasi`, `public.diklat`, `public.penghargaan`.
   - **Referensi (`ref.*`):** `agama`, `status_kepegawaian`, `status_keaktifan_pegawai`, `jenis_ptk`, `jabatan_ptk`, `jabatan_tugas_ptk`, `pangkat_golongan`, `jenjang_pendidikan`, `bidang_studi`, `gelar_akademik`, `mata_pelajaran`, `mst_wilayah`.
3. **Domain Siswa / Peserta Didik (`public.peserta_didik`):**
   - **Utama:** `public.peserta_didik` (`peserta_didik_id`), `public.registrasi_peserta_didik`, `public.anggota_rombel`, `public.peserta_didik_longitudinal` (`semester_id`), `public.beasiswa_peserta_didik`, `public.kesejahteraan_pd`, `public.prestasi`, `public.ijazah_pd`, `public.tracer_lulusan`.
   - **Referensi (`ref.*`):** `agama`, `kebutuhan_khusus`, `jenis_tinggal`, `alat_transportasi`, `pekerjaan`, `penghasilan`, `jenjang_pendidikan`, `jenis_pendaftaran`, `jenis_keluar`, `tingkat_pendidikan`, `mst_wilayah`.

---

## 3. Daftar Ceklis Pengerjaan Bertahap (Step-by-Step Checklist)

### Fase 1: Setup Modul Go, `pgxpool` Read-Only & Middleware
- [x] **1.1 Inisialisasi Go Module & Struktur Folder:** Buat modul `dapodik-readonly-api`, gunakan router standar `net/http` (alternatif `chi` yang diizinkan), tambahkan `pgx/v5` dan `godotenv`, lalu buat struktur folder di dalam `cmd/` dan `internal/`.
- [x] **1.2 Konfigurasi `pgxpool` Read-Only:** Pada `internal/database/postgres.go`, konfigurasikan `pgxpool.Config` agar setiap koneksi baru mengeksekusi `SET default_transaction_read_only = on` (atau tambahkan parameter `options=-c%20default_transaction_read_only%3Don` pada DSN PostgreSQL) serta atur *connection timeout* dan *pool size*.
- [x] **1.3 Response Helper & Pagination Utils:** Buat `internal/response/json.go` untuk format standar JSON (`Success`, `Message`, `Data`, `Meta`) dan penanganan nilai `NULL` SQL (gunakan pointer `*string`, `*int`, atau `pgtype` agar aman saat serialisasi JSON).
- [x] **1.4 Middleware Read-Only & API Key:** Buat middleware yang memblokir semua *request* selain `GET`, `HEAD`, dan `OPTIONS` dengan kode HTTP `405 Method Not Allowed`, serta middleware autentikasi `X-API-Key`.
- [x] **1.5 Verifikasi Fase 1:** Implementasikan `GET /api/health` yang melakukan `pool.Ping(ctx)` ke database `pendataan` dan pastikan `go build ./cmd/api` sukses tanpa error.

### Fase 2: Modul Referensi (`schema: ref`)
- [x] **2.1 Repository & Handler Referensi Periode:** Implementasikan:
  - `GET /api/ref/semester` (daftar semester beserta status `periode_aktif`)
  - `GET /api/ref/tahun-ajaran` (daftar tahun ajaran beserta status `periode_aktif`)
- [x] **2.2 Resolusi Wilayah Bertingkat:** Implementasikan `GET /api/ref/wilayah/{kode_wilayah}` yang melakukan *self-join* pada `ref.mst_wilayah` menggunakan `mst_kode_wilayah` untuk mengembalikan hierarki lengkap (Desa/Kelurahan, Kecamatan, Kabupaten/Kota, Provinsi).
- [x] **2.3 Verifikasi Fase 2:** Uji endpoint referensi dan pastikan query berjalan cepat serta lolos `go vet ./...`.

### Fase 3: Modul Data Sekolah (`public.sekolah`)
- [x] **3.1 Profil Utama Sekolah:** Implementasikan `GET /api/sekolah` (mendukung `?npsn=`, `?page=`, dan `?limit=`; tanpa `limit`, mengembalikan semua hasil) dan `GET /api/sekolah/{sekolah_id}` dengan SQL `JOIN` ke `ref.bentuk_pendidikan`, `ref.status_kepemilikan`, `ref.mst_wilayah`, serta `public.akreditasi_sp` terbaru (pastikan kolom `no_rekening` dan `npwp` tidak di-SELECT).
- [x] **3.2 Data Longitudinal & Utilitas Sekolah:** Implementasikan `GET /api/sekolah/{sekolah_id}/longitudinal` (menerima query param `?semester_id=`) untuk menggabungkan `public.sekolah_longitudinal`, `public.listrik`, `public.internet`, dan `public.sanitasi`.
- [x] **3.3 Jurusan & Rombongan Belajar:** Implementasikan:
  - `GET /api/sekolah/{sekolah_id}/jurusan` (`public.jurusan_sp` JOIN `ref.jurusan`)
  - `GET /api/sekolah/{sekolah_id}/rombel` (`public.rombongan_belajar` JOIN `ref.tingkat_pendidikan`, `ref.kurikulum`, `public.ruang`, dan `public.ptk` sebagai wali kelas; filter `?semester_id=`)
- [x] **3.4 Prasarana (Bangunan & Ruang):** Implementasikan `GET /api/sekolah/{sekolah_id}/ruang` (`public.ruang` JOIN `public.bangunan` dan `ref.jenis_prasarana`).
- [x] **3.5 Verifikasi Fase 3:** Kompilasi dan uji seluruh *endpoint* `/api/sekolah/*`, pastikan *null-safety* aman untuk kolom opsional.

### Fase 4: Modul Data Guru / PTK (`public.ptk`)
- [x] **4.1 Daftar & Pencarian PTK:** Implementasikan `GET /api/ptk` dengan filter `?sekolah_id=`, `?tahun_ajaran_id=`, `?jenis_ptk_id=`, `?search=` (`ILIKE` pada `nama`, `nuptk`, `nip`), dan paginasi opsional (tanpa `limit`, mengembalikan semua hasil) dari `public.ptk` JOIN `public.ptk_terdaftar`. Respons daftar dan detail PTK menyertakan `nik`, `no_kk`, `npwp`, `rekening_bank`, `rekening_atas_nama`, `karpeg`, dan `karpas` sesuai pengecualian data PTK di atas.
- [x] **4.2 Detail Biodata PTK:** Implementasikan `GET /api/ptk/{ptk_id}` yang memuat biodata kepegawaian beserta label dari `ref.agama`, `ref.status_kepegawaian`, `ref.pangkat_golongan`, `ref.jenis_ptk`, dan `ref.jabatan_ptk`.
- [x] **4.3 Beban Mengajar & Tugas Tambahan:** Implementasikan:
  - `GET /api/ptk/{ptk_id}/pembelajaran` (`public.pembelajaran` JOIN `public.rombongan_belajar` & `ref.mata_pelajaran`, filter `?semester_id=`)
  - `GET /api/ptk/{ptk_id}/tugas-tambahan` (`public.tugas_tambahan` JOIN `ref.jabatan_tugas_ptk`)
- [x] **4.4 Riwayat Pendidikan, Kepangkatan & Sertifikasi:** Implementasikan `GET /api/ptk/{ptk_id}/riwayat` untuk mengambil riwayat dari `public.rwy_pend_formal`, `public.rwy_kepangkatan`, `public.rwy_sertifikasi`, dan `public.diklat` beserta label `ref.*`.
- [x] **4.5 Verifikasi Fase 4:** Kompilasi dan uji seluruh *endpoint* `/api/ptk/*` termasuk pencarian dan filter tahun ajaran.

### Fase 5: Modul Data Siswa / Peserta Didik (`public.peserta_didik`)
- [x] **5.1 Daftar & Pencarian Siswa:** Implementasikan `GET /api/peserta-didik` khusus siswa yang registrasi terbarunya belum keluar (`jenis_keluar_id IS NULL`) dan rombel terbarunya memiliki `jenis_rombel` 1 atau 9, dengan relasi `anggota_rombel.rombongan_belajar_id` ke `rombongan_belajar.rombongan_belajar_id`. Dukung filter `?sekolah_id=`, `?rombongan_belajar_id=`, `?semester_id=`, `?tingkat_pendidikan_id=`, `?search=` (`nama`, `nisn`, `nipd`), dan paginasi opsional (tanpa `page`/`limit`, kembalikan semua hasil). Status keluar tidak ditampilkan pada item daftar; kolom identitas sensitif siswa dan keluarga tidak diambil.
- [x] **5.2 Detail Profil & Data Keluarga Siswa:** Implementasikan `GET /api/peserta-didik/{peserta_didik_id}` yang menggabungkan `public.peserta_didik`, `public.registrasi_peserta_didik`, label referensi (`ref.agama`, `ref.jenis_tinggal`, `ref.alat_transportasi`, `ref.pekerjaan`, `ref.penghasilan`), serta data tinggi/berat/jarak terbaru dari `public.peserta_didik_longitudinal`. Kolom NIK/KK, identitas NIK keluarga, rekening, dan nomor bantuan tidak diambil.
- [x] **5.3 Riwayat Rombel & Prestasi/Beasiswa:** Implementasikan:
- [x] `GET /api/peserta-didik/{peserta_didik_id}/rombel` (`public.anggota_rombel` JOIN `public.rombongan_belajar`, `ref.semester`, dan `public.ptk` wali kelas)
- [x] `GET /api/peserta-didik/{peserta_didik_id}/prestasi-beasiswa` (`public.prestasi` & `public.beasiswa_peserta_didik` JOIN tabel `ref` terkait)
- [x] **5.4 Build Binary & Dokumentasi Akhir:**
  - [x] Pastikan `go build -o bin/dapodik-api ./cmd/api` berjalan bersih tanpa *warning*.
  - [x] Buat file dokumentasi `API_DOCS.md` berisi daftar lengkap *endpoint*, contoh `curl`, query parameters, dan contoh JSON response.
### Fase 6: Modul Data Kelas / Rombongan Belajar (`public.rombongan_belajar`)

* [x] **6.1 Daftar & Pencarian Rombel:** Implementasikan `GET /api/rombel` dengan dukungan filter `?sekolah_id=`, `?semester_id=`, `?tingkat_pendidikan_id=`, `?jenis_rombel=`, `?kurikulum_id=`, `?jurusan_sp_id=`, `?ptk_id=` (wali kelas), `?search=` (`ILIKE` pada `nama`), serta paginasi opsional (tanpa `limit`, mengembalikan semua hasil). Jenis rombel `18` dan `51` dikecualikan. Sertakan agregasi jumlah siswa aktif dan label relasi sekolah, wali kelas, ruang, semester, tingkat, kurikulum, jenis rombel, dan jurusan.

* [x] **6.2 Detail Lengkap Rombel & Ekstrakurikuler:** Implementasikan `GET /api/rombel/{rombongan_belajar_id}` yang memuat:

  * Metadata lengkap rombel beserta label dari `ref.semester`, `ref.tingkat_pendidikan`, `ref.kurikulum`, `ref.jenis_rombel`, dan `ref.kebutuhan_khusus`.

  * Informasi wali kelas (`public.ptk` — tanpa kolom sensitif) & prasarana ruang kelas (`public.ruang` JOIN `public.bangunan`).

  * Informasi jurusan (`public.jurusan_sp` JOIN `ref.jurusan`) jika rombel memiliki kompetensi keahlian.

  * Informasi ekstrakurikuler apabila rombel bertipe ekstrakurikuler. **Skema lokal:** `public.ekstrakurikuler` dan `ref.ekskul` belum tersedia; respons menandai `ekstrakurikuler_tersedia: false`.

  * Status validasi lokal Dapodik (`LEFT JOIN public.vld_rombel`) jika tersedia.

* [x] **6.3 Daftar Anggota Kelas (Siswa dalam Rombel):** Implementasikan `GET /api/rombel/{rombongan_belajar_id}/anggota` dengan label pendaftaran, agama, dan status keluar. Kolom sensitif siswa dikecualikan; urutan alfabetis (`ASC`).

* [x] **6.4 Daftar Pembelajaran, Mata Pelajaran & Guru Pengampu:** Implementasikan `GET /api/rombel/{rombongan_belajar_id}/pembelajaran` yang menggabungkan mapel dan guru untuk menampilkan nama mapel, guru, SK/tanggal SK, jam per minggu, serta status induk/non-induk. **Skema lokal:** kolom kode mapel tidak tersedia; `kode_mata_pelajaran` bernilai null.

* [x] **6.5 Rekapitulasi Statistik & Nilai/Evaluasi Rombel:** Implementasikan `GET /api/rombel/{rombongan_belajar_id}/statistik` berisi total siswa, jumlah L/P, sebaran agama dan usia, serta jumlah pembelajaran/mapel. **Skema lokal:** `public.matev_rapor` dan `public.nilai_rapor` belum tersedia; `evaluasi_rapor_tersedia` bernilai false.

* [x] **6.6 Verifikasi Fase 6, Build Binary & Dokumentasi Akhir:**

  * [x] Uji seluruh *endpoint* `/api/rombel/*` terhadap nilai `jenis_rombel` yang tersedia di database dan pastikan `go vet ./...` bersih.

  * [x] Pastikan build binary API berjalan bersih tanpa *warning*.

  * Buat/perbarui file dokumentasi `API_DOCS.md` berisi daftar lengkap seluruh *endpoint* (Fase 1–6), contoh `curl`, query parameters, dan contoh JSON response.
