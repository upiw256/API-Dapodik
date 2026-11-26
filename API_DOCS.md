# Dapodik Read-Only API

Semua endpoint hanya menerima `GET` dan membutuhkan header `X-API-Key`. Ganti `ISI_API_KEY` dengan nilai `API_KEY` dari `.env`; jangan membagikan key tersebut.

Atur `SEMESTER` di `.env`, misalnya `SEMESTER="2026/2027 Gasal"`. Saat startup aplikasi mencocokkan label tersebut dengan `ref.semester`; API gagal startup jika nilainya tidak ditemukan atau ambigu. Seluruh endpoint yang datanya terikat semester memakai semester pilihan ini, termasuk daftar/detail kelas dan siswa, longitudinal sekolah, serta pembelajaran PTK. Parameter `semester_id`/`tahun_ajaran_id` pada request tidak dapat mengganti pilihan `.env`. Endpoint yang secara alami berupa profil, referensi wilayah, atau catatan historis tanpa relasi semester tetap menampilkan datanya sesuai tabel.

Untuk endpoint daftar yang menyediakan `limit`, hasil tidak dipotong jika `limit` tidak dikirim. `page` hanya diterapkan bersama `limit`. Jika `limit` diberikan tanpa `page`, halaman pertama digunakan. Nilai maksimum `limit` adalah 100.

Rate limit berlaku gabungan pada semua endpoint: maksimal 100 request terautentikasi per alamat IP. Request berikutnya mendapat HTTP `429 Too Many Requests` dan header `Retry-After` dalam detik. Jika tidak ada request selama 1 menit, hitungan di-reset walaupun jendela 10 menit belum lewat. Request `OPTIONS` untuk preflight tidak dihitung. Jika API berada di belakang reverse proxy, batas ini menggunakan alamat peer yang dilihat aplikasi.

```powershell
$headers = @{ "X-API-Key" = "ISI_API_KEY" }
$baseUrl = "http://localhost:8080"
```

## Referensi

| Method | Endpoint | Parameter |
| --- | --- | --- |
| GET | `/api/ref/semester` | — |
| GET | `/api/ref/tahun-ajaran` | — |
| GET | `/api/ref/wilayah/{kode_wilayah}` | Kode wilayah pada path |

## Sekolah

| Method | Endpoint | Parameter |
| --- | --- | --- |
| GET | `/api/sekolah` | `npsn`, `page`, `limit` |
| GET | `/api/sekolah/detail` | — |
| GET | `/api/sekolah/longitudinal` | `semester_id` |
| GET | `/api/sekolah/jurusan` | — |
| GET | `/api/sekolah/rombel` | `semester_id` |
| GET | `/api/sekolah/ruang` | — |

Atur `NPSN` di `.env` untuk menentukan sekolah yang dipakai seluruh endpoint ini; aplikasi gagal startup jika `NPSN` belum diisi. Daftar `/api/sekolah` (dengan atau tanpa garis miring di akhir URL) selalu dibatasi ke NPSN tersebut; parameter `npsn` tidak dapat memperluas hasil. Endpoint detail dan data terkait tidak memerlukan UUID sekolah karena aplikasi mengambil sekolah dari konfigurasi `NPSN`.

## PTK

| Method | Endpoint | Parameter |
| --- | --- | --- |
| GET | `/api/ptk` | `sekolah_id`, `tahun_ajaran_id`, `jenis_ptk_id`, `search`, `page`, `limit` |
| GET | `/api/ptk/{ptk_id}` | UUID PTK |
| GET | `/api/ptk/{ptk_id}/pembelajaran` | `semester_id` |
| GET | `/api/ptk/{ptk_id}/tugas-tambahan` | — |
| GET | `/api/ptk/{ptk_id}/riwayat` | — |

Data PTK mencakup pengecualian yang dirinci dalam implementation plan. Jangan membagikan respons yang berisi data pribadi.

## Peserta didik

| Method | Endpoint | Parameter |
| --- | --- | --- |
| GET | `/api/peserta-didik` | `sekolah_id`, `rombongan_belajar_id`, `semester_id`, `tingkat_pendidikan_id`, `search`, `page`, `limit` |
| GET | `/api/peserta-didik/{peserta_didik_id}` | UUID peserta didik |
| GET | `/api/peserta-didik/{peserta_didik_id}/rombel` | UUID peserta didik |
| GET | `/api/peserta-didik/{peserta_didik_id}/prestasi-beasiswa` | UUID peserta didik |

Endpoint daftar hanya mengembalikan siswa yang registrasi terbarunya belum mencatat keluar (`jenis_keluar_id IS NULL`) dan berada pada rombel terbaru jenis `1` atau `9`. Relasi siswa-kelas menggunakan `rombongan_belajar_id` melalui `anggota_rombel`. `jenis_rombel` ikut ditampilkan, sedangkan status keluar tidak. `search` mencari nama, NISN, atau NIPD. Kolom NIK/KK siswa dan keluarga, rekening, serta nomor bantuan tidak disertakan.

Contoh pemanggilan dan filter gabungan di PowerShell:

```powershell
Invoke-RestMethod -Headers $headers -Uri "$baseUrl/api/peserta-didik?sekolah_id=UUID-SEKOLAH&search=Contoh&page=1&limit=10"
Invoke-RestMethod -Headers $headers -Uri "$baseUrl/api/peserta-didik/UUID-PESERTA-DIDIK"
Invoke-RestMethod -Headers $headers -Uri "$baseUrl/api/peserta-didik/UUID-PESERTA-DIDIK/rombel"
Invoke-RestMethod -Headers $headers -Uri "$baseUrl/api/peserta-didik/UUID-PESERTA-DIDIK/prestasi-beasiswa"
```

## Bentuk respons

Respons sukses memakai envelope umum berikut. Nilai di bawah hanya contoh ilustratif.

```json
{
  "success": true,
  "message": "daftar peserta didik berhasil diambil",
  "data": [
    {
      "peserta_didik_id": "UUID-PESERTA-DIDIK",
      "nama": "Nama Contoh",
      "jenis_kelamin": "L",
      "sekolah_id": "UUID-SEKOLAH"
    }
  ],
  "meta": { "page": 1, "limit": 20, "total": 1, "total_pages": 1 }
}
```

Respons detail menyediakan `nama_peserta_didik` setelah `message`, serta biodata dan label referensi, data keluarga non-identifikator, dan tinggi/berat/jarak dari data longitudinal pada semester pilihan. Endpoint rombel peserta didik hanya mengembalikan kelas pada semester pilihan; endpoint prestasi/beasiswa mengembalikan objek dengan array `prestasi` dan `beasiswa`.

## Rombongan belajar (rombel)

| Method | Endpoint | Parameter |
| --- | --- | --- |
| GET | `/api/rombel` | `sekolah_id`, `semester_id`, `tingkat_pendidikan_id`, `jenis_rombel`, `kurikulum_id`, `jurusan_sp_id`, `ptk_id`, `search`, `page`, `limit` |
| GET | `/api/pembelajaran` | Semua pembelajaran semester aktif, termasuk ID rombel pada tiap baris |
| GET | `/api/rombel/{rombongan_belajar_id}` | UUID rombel |
| GET | `/api/rombel/{rombongan_belajar_id}/anggota` | UUID rombel |
| GET | `/api/rombel/{rombongan_belajar_id}/pembelajaran` | UUID rombel |
| GET | `/api/rombel/{rombongan_belajar_id}/statistik` | UUID rombel |

Daftar rombel menyertakan jumlah siswa aktif dan label sekolah, wali kelas, ruang, semester, tingkat, kurikulum, jenis rombel, serta jurusan jika tersedia. Jenis rombel `18` dan `51` dikecualikan dari endpoint daftar. Detail juga berisi kebutuhan khusus dan ringkasan status validasi. Endpoint `/anggota` membungkus hasil dalam objek yang memuat `nama_kelas`, `wali_kelas`, dan array `anggota`; array tetap kosong (`[]`) jika belum ada siswa.

Endpoint `/api/pembelajaran` mengambil semua pembelajaran semester aktif dalam satu permintaan dan menyertakan `rombongan_belajar_id`, `nama_kelas`, serta `wali_kelas` pada setiap baris. Endpoint `/api/rombel/{rombongan_belajar_id}/pembelajaran` tetap tersedia untuk satu rombel dan menyertakan metadata kelas di tingkat teratas.

```powershell
Invoke-RestMethod -Headers $headers -Uri "$baseUrl/api/rombel?sekolah_id=UUID-SEKOLAH&semester_id=20251&page=1&limit=20"
Invoke-RestMethod -Headers $headers -Uri "$baseUrl/api/pembelajaran"
Invoke-RestMethod -Headers $headers -Uri "$baseUrl/api/rombel/UUID-ROMBEL"
Invoke-RestMethod -Headers $headers -Uri "$baseUrl/api/rombel/UUID-ROMBEL/anggota"
Invoke-RestMethod -Headers $headers -Uri "$baseUrl/api/rombel/UUID-ROMBEL/pembelajaran"
Invoke-RestMethod -Headers $headers -Uri "$baseUrl/api/rombel/UUID-ROMBEL/statistik"
```

Skema database yang diperiksa belum memiliki tabel ekstrakurikuler (`public.ekstrakurikuler`, `ref.ekskul`) atau tabel evaluasi rapor (`public.matev_rapor`, `public.nilai_rapor`). Detail menyatakan `ekstrakurikuler_tersedia: false`, statistik menyatakan `evaluasi_rapor_tersedia: false`, dan kode mapel bernilai null karena kolom kode tidak tersedia.

## Kontrak respons semua endpoint

Contoh di bawah memakai nilai ilustrasi. Field opsional yang bernilai kosong dapat tidak muncul karena memakai `omitempty`. Endpoint daftar yang dipaginasi mengembalikan `meta`; bila tanpa `limit`, isi `meta.limit` adalah jumlah total data.

| Endpoint | `message` sukses | Bentuk `data` dan `meta` |
| --- | --- | --- |
| `GET /api/health` | `healthy` | Objek `{ "database": "connected" }`; tanpa `meta` |
| `GET /api/ref/semester` | `data semester berhasil diambil` | Array objek semester; tanpa `meta` |
| `GET /api/ref/tahun-ajaran` | `data tahun ajaran berhasil diambil` | Array objek tahun ajaran; tanpa `meta` |
| `GET /api/ref/wilayah/{kode_wilayah}` | `hierarki wilayah berhasil diambil` | Objek dengan `desa_kelurahan`, `kecamatan`, `kabupaten_kota`, `provinsi`; tanpa `meta` |
| `GET /api/sekolah` | `daftar sekolah berhasil diambil` | Array profil sekolah; `meta`: page, limit, total, total_pages |
| `GET /api/sekolah/detail` | `detail sekolah berhasil diambil` | Satu profil sekolah; tanpa `meta` |
| `GET /api/sekolah/longitudinal` | `data longitudinal sekolah berhasil diambil` | Array data longitudinal; tanpa `meta` |
| `GET /api/sekolah/jurusan` | `jurusan sekolah berhasil diambil` | Array jurusan; tanpa `meta` |
| `GET /api/sekolah/rombel` | `rombongan belajar sekolah berhasil diambil` | Array rombel; tanpa `meta` |
| `GET /api/sekolah/ruang` | `ruang sekolah berhasil diambil` | Array ruang; tanpa `meta` |
| `GET /api/ptk` | `daftar PTK berhasil diambil` | Array PTK; `meta`: page, limit, total, total_pages |
| `GET /api/ptk/{ptk_id}` | `detail PTK berhasil diambil` | Satu detail PTK; tanpa `meta` |
| `GET /api/ptk/{ptk_id}/pembelajaran` | `data pembelajaran PTK berhasil diambil` | Array pembelajaran PTK; tanpa `meta` |
| `GET /api/ptk/{ptk_id}/tugas-tambahan` | `tugas tambahan PTK berhasil diambil` | Array tugas tambahan; tanpa `meta` |
| `GET /api/ptk/{ptk_id}/riwayat` | `riwayat PTK berhasil diambil` | Objek berisi `pendidikan_formal`, `kepangkatan`, `sertifikasi`, `diklat`; tanpa `meta` |
| `GET /api/peserta-didik` | `daftar peserta didik berhasil diambil` | Array siswa; `meta`: page, limit, total, total_pages |
| `GET /api/peserta-didik/{peserta_didik_id}` | `detail peserta didik berhasil diambil` | Detail siswa di `data`; `nama_peserta_didik` berada di tingkat teratas setelah `message`; tanpa `meta` |
| `GET /api/peserta-didik/{peserta_didik_id}/rombel` | `riwayat rombel peserta didik berhasil diambil` | Array rombel siswa; tanpa `meta` |
| `GET /api/peserta-didik/{peserta_didik_id}/prestasi-beasiswa` | `prestasi dan beasiswa peserta didik berhasil diambil` | Objek berisi array `prestasi` dan `beasiswa`; tanpa `meta` |
| `GET /api/rombel` | `daftar rombel berhasil diambil` | Array rombel; `meta`: page, limit, total, total_pages |
| `GET /api/pembelajaran` | `seluruh pembelajaran berhasil diambil` | Array pembelajaran semester aktif; setiap baris menyertakan ID dan nama rombel; tanpa `meta` |
| `GET /api/rombel/{rombongan_belajar_id}` | `detail rombel berhasil diambil` | Satu detail rombel; tanpa `meta` |
| `GET /api/rombel/{rombongan_belajar_id}/anggota` | `anggota rombel berhasil diambil` | Objek berisi identitas kelas, wali kelas, dan array `anggota`; tanpa `meta` |
| `GET /api/rombel/{rombongan_belajar_id}/pembelajaran` | `pembelajaran rombel berhasil diambil` | `nama_kelas`, `wali_kelas`, dan array pembelajaran `data` berada di tingkat teratas; tanpa `meta` |
| `GET /api/rombel/{rombongan_belajar_id}/statistik` | `statistik rombel berhasil diambil` | Objek statistik kelas; tanpa `meta` |

Respons sukses memakai `success: true`. Contoh bentuk spesifik tiap kelompok:

Untuk endpoint detail sekolah, `data` memakai objek profil sekolah yang sama dengan satu item pada `/api/sekolah`. Untuk endpoint detail rombel, `data` berupa objek detail rombel. Untuk detail PTK, `data` memakai objek detail PTK; seluruh field personal mengikuti isi database, jadi contoh di dokumen sengaja tidak menampilkan nilai asli.

### Health dan referensi

```json
{"success":true,"message":"healthy","data":{"database":"connected"}}
```

```json
{"success":true,"message":"data semester berhasil diambil","data":[{"semester_id":"20261","tahun_ajaran_id":2026,"nama":"2026/2027 Gasal","semester":1,"periode_aktif":1,"tanggal_mulai":"2026-07-01","tanggal_selesai":"2026-12-31"}]}
```

```json
{"success":true,"message":"data tahun ajaran berhasil diambil","data":[{"tahun_ajaran_id":2026,"nama":"2026/2027","periode_aktif":1,"tanggal_mulai":"2026-07-01","tanggal_selesai":"2027-06-30"}]}
```

```json
{"success":true,"message":"hierarki wilayah berhasil diambil","data":{"desa_kelurahan":{"kode_wilayah":"3273010001","nama":"Contoh Kelurahan"},"kecamatan":{"kode_wilayah":"327301","nama":"Contoh Kecamatan"},"kabupaten_kota":{"kode_wilayah":"3273","nama":"Contoh Kota"},"provinsi":{"kode_wilayah":"32","nama":"Contoh Provinsi"}}}
```

### Sekolah

Semua array endpoint sekolah berisi objek sesuai jenis datanya; berikut contoh field utamanya:

```json
{"success":true,"message":"daftar sekolah berhasil diambil","data":[{"sekolah_id":"UUID-SEKOLAH","npsn":"20227907","nama":"Nama Sekolah","bentuk_pendidikan_id":13,"status_sekolah":1,"status_kepemilikan_id":1,"alamat_jalan":"Jalan Contoh","desa_kelurahan":"Kelurahan Contoh","kode_wilayah":"3273010001","luas_tanah_milik":3,"luas_tanah_bukan_milik":0}],"meta":{"page":1,"limit":20,"total":1,"total_pages":1}}
```

```json
{"success":true,"message":"detail sekolah berhasil diambil","data":{"sekolah_id":"UUID-SEKOLAH","npsn":"20227907","nama":"Nama Sekolah","bentuk_pendidikan_id":13,"status_sekolah":1,"status_kepemilikan_id":1,"alamat_jalan":"Jalan Contoh","desa_kelurahan":"Kelurahan Contoh","kode_wilayah":"3273010001","luas_tanah_milik":3,"luas_tanah_bukan_milik":0}}
```

```json
{"success":true,"message":"data longitudinal sekolah berhasil diambil","data":[{"sekolah_id":"UUID-SEKOLAH","semester_id":"20261","waktu_penyelenggaraan_id":1,"kontinuitas_listrik":"Menerus","wilayah_terpencil":0,"wilayah_perbatasan":0,"wilayah_transmigrasi":0,"wilayah_adat_terpencil":0,"wilayah_bencana_alam":0,"wilayah_bencana_sosial":0,"partisipasi_bos":1,"sertifikasi_iso_id":0,"sumber_listrik_id":1,"daya_listrik":2200,"akses_internet_2_id":0,"listrik":[],"internet":[]} ]}
```

```json
{"success":true,"message":"jurusan sekolah berhasil diambil","data":[{"jurusan_sp_id":"UUID-JURUSAN-SP","jurusan_id":"UUID-JURUSAN","nama_jurusan_sp":"IPA"}]}
```

```json
{"success":true,"message":"rombongan belajar sekolah berhasil diambil","data":[{"rombongan_belajar_id":"UUID-ROMBEL","semester_id":"20261","nama":"X IPA 1","tingkat_pendidikan_id":10,"kurikulum_id":1,"tanggal_mulai":"2026-07-14","tanggal_selesai":"2026-12-31","jenis_rombel":1}]}
```

```json
{"success":true,"message":"ruang sekolah berhasil diambil","data":[{"ruang_id":"UUID-RUANG","jenis_prasarana_id":1,"kode_ruang":"R-01","nama_ruang":"Kelas X IPA 1","lantai":1}]}
```

### PTK

```json
{"success":true,"message":"daftar PTK berhasil diambil","data":[{"ptk_id":"UUID-PTK","nama":"Nama PTK","nip":"198001012005011001","nik":"0000000000000000","jenis_kelamin":"L","status_kepegawaian_id":1,"sekolah_id":"UUID-SEKOLAH","tahun_ajaran_id":2026,"jenis_ptk_id":3}],"meta":{"page":1,"limit":20,"total":1,"total_pages":1}}
```

`GET /api/ptk/{ptk_id}` mengembalikan satu objek detail PTK. Field sensitif seperti NIK, nomor KK, NPWP, rekening, dan nomor kartu dapat muncul sesuai data sumber; jangan membagikan respons detail.

```json
{"success":true,"message":"detail PTK berhasil diambil","data":{"ptk_id":"UUID-PTK","nama":"Nama PTK","nik":"0000000000000000","jenis_kelamin":"L","tempat_lahir":"Kota Contoh","tanggal_lahir":"1980-01-01","agama_id":1,"status_kepegawaian_id":1,"status_keaktifan_id":1,"alamat_jalan":"Jalan Contoh","desa_kelurahan":"Kelurahan Contoh","kode_wilayah":"3273010001"}}
```

```json
{"success":true,"message":"data pembelajaran PTK berhasil diambil","data":[{"pembelajaran_id":"UUID-PEMBELAJARAN","semester_id":"20261","rombongan_belajar_id":"UUID-ROMBEL","rombel":"X IPA 1","mata_pelajaran_id":300110000,"mata_pelajaran":"Bahasa Indonesia","sk_mengajar":"936/PK.02.01/SMAN1","tanggal_sk_mengajar":"2026-07-14","jam_mengajar_per_minggu":4,"status_di_kurikulum":1}]}
```

```json
{"success":true,"message":"tugas tambahan PTK berhasil diambil","data":[{"ptk_tugas_tambahan_id":"UUID-TUGAS","jabatan_ptk_id":1,"sekolah_id":"UUID-SEKOLAH","jumlah_jam":2,"nomor_sk":"SK/001","tmt_tambahan":"2026-07-01"}]}
```

```json
{"success":true,"message":"riwayat PTK berhasil diambil","data":{"pendidikan_formal":[],"kepangkatan":[],"sertifikasi":[],"diklat":[]}}
```

### Peserta didik

```json
{"success":true,"message":"daftar peserta didik berhasil diambil","data":[{"peserta_didik_id":"UUID-PESERTA-DIDIK","nama":"Nama Siswa","jenis_kelamin":"P","nisn":"0000000000","nipd":"10001","sekolah_id":"UUID-SEKOLAH","rombongan_belajar_id":"UUID-ROMBEL","rombel":"X IPA 1","jenis_rombel":1,"semester_id":"20261","tingkat_pendidikan_id":10,"tingkat_pendidikan":"10"}],"meta":{"page":1,"limit":20,"total":1,"total_pages":1}}
```

```json
{"success":true,"message":"detail peserta didik berhasil diambil","nama_peserta_didik":"Nama Siswa","data":{"peserta_didik_id":"UUID-PESERTA-DIDIK","nama":"Nama Siswa","jenis_kelamin":"P"}}
```

```json
{"success":true,"message":"riwayat rombel peserta didik berhasil diambil","data":[{"anggota_rombel_id":"UUID-ANGGOTA","rombongan_belajar_id":"UUID-ROMBEL","nama":"X IPA 1","semester_id":"20261","sekolah_id":"UUID-SEKOLAH"}]}
```

```json
{"success":true,"message":"prestasi dan beasiswa peserta didik berhasil diambil","data":{"prestasi":[],"beasiswa":[]}}
```

### Rombongan belajar

```json
{"success":true,"message":"daftar rombel berhasil diambil","data":[{"rombongan_belajar_id":"UUID-ROMBEL","nama":"X IPA 1","semester_id":"20261","sekolah_id":"UUID-SEKOLAH","tingkat_pendidikan_id":10,"jenis_rombel":1,"kurikulum_id":1,"jumlah_siswa_aktif":30}],"meta":{"page":1,"limit":20,"total":1,"total_pages":1}}
```

```json
{"success":true,"message":"detail rombel berhasil diambil","data":{"rombongan_belajar_id":"UUID-ROMBEL","nama":"X IPA 1","semester_id":"20261","sekolah_id":"UUID-SEKOLAH","tingkat_pendidikan_id":10,"kurikulum_id":1,"jenis_rombel":1,"jumlah_siswa_aktif":30,"jumlah_validasi":0,"ekstrakurikuler_tersedia":false}}
```

```json
{"success":true,"message":"anggota rombel berhasil diambil","data":{"rombongan_belajar_id":"UUID-ROMBEL","nama_kelas":"X IPA 1","wali_kelas":"Nama PTK","anggota":[]}}
```

```json
{"success":true,"message":"pembelajaran rombel berhasil diambil","nama_kelas":"X IPA 1","wali_kelas":"Nama PTK","data":[{"pembelajaran_id":"UUID-PEMBELAJARAN","semester_id":"20261","mata_pelajaran_id":300110000,"mata_pelajaran":"Bahasa Indonesia","ptk_id":"UUID-PTK","nama_guru":"Nama PTK","ptk_terdaftar_id":"UUID-PTK-TERDAFTAR","sk_mengajar":"SK/001","jam_mengajar_per_minggu":4,"status_di_kurikulum":1,"induk":true}]}
```

```json
{"success":true,"message":"seluruh pembelajaran berhasil diambil","data":[{"pembelajaran_id":"UUID-PEMBELAJARAN","rombongan_belajar_id":"UUID-ROMBEL","nama_kelas":"X IPA 1","wali_kelas":"Nama Wali Kelas","semester_id":"20261","mata_pelajaran_id":300110000,"mata_pelajaran":"Bahasa Indonesia","ptk_id":"UUID-PTK-PENGAJAR","nama_guru":"Nama Pengajar","ptk_terdaftar_id":"UUID-PTK-TERDAFTAR","sk_mengajar":"SK/001","jam_mengajar_per_minggu":4,"status_di_kurikulum":1,"induk":true}]}
```

```json
{"success":true,"message":"statistik rombel berhasil diambil","data":{"total_siswa":30,"laki_laki":14,"perempuan":16,"jenis_kelamin_tidak_diketahui":0,"sebaran_agama":[],"sebaran_usia":[],"jumlah_pembelajaran":10,"jumlah_mata_pelajaran":8,"evaluasi_rapor_tersedia":false}}
```

Respons gagal memakai bentuk umum berikut; pesan menyesuaikan endpoint. UUID atau parameter tidak valid menghasilkan `400`, data detail tidak ditemukan menghasilkan `404`, dan request yang melewati rate limit menghasilkan `429`.

```json
{"success":false,"message":"peserta didik tidak ditemukan"}
```

## Menjalankan saat boot tanpa login

Build binary Windows dari komputer pengembangan, lalu salin `dapodik-api.exe`, `.env`, dan `install-service.ps1` ke folder yang sama di server. Buka PowerShell **Run as administrator**, pindah ke folder tersebut, lalu jalankan:

```powershell
.\install-service.ps1
```

Script mendaftarkan Scheduled Task `DapodikAPI` dengan trigger saat startup. Task berjalan sebagai akun `NT AUTHORITY\SYSTEM`, jadi tidak menunggu pengguna login. Working directory task adalah folder binary; aplikasi membaca `.env` dari folder tersebut. Jika task sudah ada, script menggantinya, lalu langsung menjalankan task. Windows akan mencoba memulai ulang proses hingga tiga kali dengan jeda satu menit jika proses berhenti.

Jika server sebelumnya menjalankan Windows Service lama bernama `DapodikAPI`, hentikan dan nonaktifkan service lama sebelum menjalankan Scheduled Task agar dua proses tidak berebut port `8080`:

```powershell
Stop-Service DapodikAPI
Set-Service DapodikAPI -StartupType Disabled
```

Periksa task dan hasil eksekusi terakhir:

```powershell
Get-ScheduledTask -TaskName DapodikAPI | Select-Object TaskName, State
Get-ScheduledTaskInfo -TaskName DapodikAPI | Select-Object LastRunTime, LastTaskResult, NextRunTime
```

Untuk memulai ulang task secara manual:

```powershell
Stop-ScheduledTask -TaskName DapodikAPI
Start-ScheduledTask -TaskName DapodikAPI
```

Untuk melepas task (binary dan `.env` tetap ada), jalankan PowerShell Administrator:

```powershell
.\uninstall-service.ps1
```

Kesalahan validasi query/path menggunakan HTTP `400`; data detail yang tidak ditemukan menggunakan `404`; kesalahan server menggunakan `500` dengan pesan umum.
