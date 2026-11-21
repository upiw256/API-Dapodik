# Dapodik Read-Only API

Semua endpoint hanya menerima `GET` dan membutuhkan header `X-API-Key`. Ganti `ISI_API_KEY` dengan nilai `API_KEY` dari `.env`; jangan membagikan key tersebut.

Untuk endpoint daftar yang menyediakan `limit`, hasil tidak dipotong jika `limit` tidak dikirim. `page` hanya diterapkan bersama `limit`. Jika `limit` diberikan tanpa `page`, halaman pertama digunakan. Nilai maksimum `limit` adalah 100.

Rate limit berlaku gabungan pada semua endpoint: maksimal 5 request terautentikasi per alamat IP. Request berikutnya mendapat HTTP `429 Too Many Requests` dan header `Retry-After` dalam detik. Jika tidak ada request selama 1 menit, hitungan di-reset walaupun jendela 10 menit belum lewat. Request `OPTIONS` untuk preflight tidak dihitung. Jika API berada di belakang reverse proxy, batas ini menggunakan alamat peer yang dilihat aplikasi.

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
| GET | `/api/sekolah/{sekolah_id}` | UUID sekolah |
| GET | `/api/sekolah/{sekolah_id}/longitudinal` | `semester_id` |
| GET | `/api/sekolah/{sekolah_id}/jurusan` | — |
| GET | `/api/sekolah/{sekolah_id}/rombel` | `semester_id` |
| GET | `/api/sekolah/{sekolah_id}/ruang` | — |

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
| GET | `/api/peserta-didik` | `sekolah_id`, `rombongan_belajar_id`, `semester_id`, `tingkat_pendidikan_id`, `aktif`, `search`, `page`, `limit` |
| GET | `/api/peserta-didik/{peserta_didik_id}` | UUID peserta didik |
| GET | `/api/peserta-didik/{peserta_didik_id}/rombel` | UUID peserta didik |
| GET | `/api/peserta-didik/{peserta_didik_id}/prestasi-beasiswa` | UUID peserta didik |

`aktif=true` memilih siswa dengan `jenis_keluar_id IS NULL`; `aktif=false` memilih siswa yang memiliki `jenis_keluar_id`. `search` mencari nama, NISN, atau NIPD. Kolom NIK/KK siswa dan keluarga, rekening, serta nomor bantuan tidak disertakan.

Contoh pemanggilan dan filter gabungan di PowerShell:

```powershell
Invoke-RestMethod -Headers $headers -Uri "$baseUrl/api/peserta-didik?sekolah_id=UUID-SEKOLAH&aktif=true&search=Contoh&page=1&limit=10"
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

Respons detail menyediakan biodata dan label referensi, data keluarga non-identifikator, serta tinggi/berat/jarak dari data longitudinal terbaru. Endpoint rombel mengembalikan riwayat kelas; endpoint prestasi/beasiswa mengembalikan objek dengan array `prestasi` dan `beasiswa`.

## Rombongan belajar (rombel)

| Method | Endpoint | Parameter |
| --- | --- | --- |
| GET | `/api/rombel` | `sekolah_id`, `semester_id`, `tingkat_pendidikan_id`, `jenis_rombel`, `kurikulum_id`, `jurusan_sp_id`, `ptk_id`, `search`, `page`, `limit` |
| GET | `/api/rombel/{rombongan_belajar_id}` | UUID rombel |
| GET | `/api/rombel/{rombongan_belajar_id}/anggota` | UUID rombel |
| GET | `/api/rombel/{rombongan_belajar_id}/pembelajaran` | UUID rombel |
| GET | `/api/rombel/{rombongan_belajar_id}/statistik` | UUID rombel |

Daftar rombel menyertakan jumlah siswa aktif dan label sekolah, wali kelas, ruang, semester, tingkat, kurikulum, jenis rombel, serta jurusan jika tersedia. Detail juga berisi kebutuhan khusus dan ringkasan status validasi.

```powershell
Invoke-RestMethod -Headers $headers -Uri "$baseUrl/api/rombel?sekolah_id=UUID-SEKOLAH&semester_id=20251&page=1&limit=20"
Invoke-RestMethod -Headers $headers -Uri "$baseUrl/api/rombel/UUID-ROMBEL"
Invoke-RestMethod -Headers $headers -Uri "$baseUrl/api/rombel/UUID-ROMBEL/anggota"
Invoke-RestMethod -Headers $headers -Uri "$baseUrl/api/rombel/UUID-ROMBEL/pembelajaran"
Invoke-RestMethod -Headers $headers -Uri "$baseUrl/api/rombel/UUID-ROMBEL/statistik"
```

Skema database yang diperiksa belum memiliki tabel ekstrakurikuler (`public.ekstrakurikuler`, `ref.ekskul`) atau tabel evaluasi rapor (`public.matev_rapor`, `public.nilai_rapor`). Detail menyatakan `ekstrakurikuler_tersedia: false`, statistik menyatakan `evaluasi_rapor_tersedia: false`, dan kode mapel bernilai null karena kolom kode tidak tersedia.

## Menjalankan sebagai Windows Service

Build binary Windows dari komputer pengembangan, lalu salin `dapodik-api.exe`, `.env`, dan `install-service.ps1` ke folder yang sama di server. Buka PowerShell **Run as administrator**, pindah ke folder tersebut, lalu jalankan:

```powershell
.\install-service.ps1
```

Script mendaftarkan service bernama `DapodikAPI` dengan startup otomatis, mengaktifkannya, dan mengatur Windows untuk mencoba restart service jika proses berhenti. Binary membaca `.env` di foldernya, meskipun working directory service adalah `C:\Windows\System32`. Log service ditulis ke `dapodik-api.log` di folder binary.

```powershell
Get-Service DapodikAPI
Restart-Service DapodikAPI
Get-Content .\dapodik-api.log -Tail 50
```

Untuk melepas service (binary dan `.env` tetap ada), jalankan PowerShell Administrator:

```powershell
.\uninstall-service.ps1
```

Kesalahan validasi query/path menggunakan HTTP `400`; data detail yang tidak ditemukan menggunakan `404`; kesalahan server menggunakan `500` dengan pesan umum.
