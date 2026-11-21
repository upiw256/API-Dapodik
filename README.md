# Dapodik Read-Only API

REST API berbasis Go untuk membaca data Dapodik melalui PostgreSQL. Endpoint API hanya menerima operasi baca. Lihat [API_DOCS.md](API_DOCS.md) untuk daftar endpoint dan contoh request.

## Kebutuhan

- Go 1.22 atau lebih baru untuk development/build.
- PostgreSQL yang dapat dijangkau aplikasi.
- File `.env` dengan `DB_DSN` dan `API_KEY`.

## Menjalankan untuk development

Salin `.env.example` menjadi `.env`, isi koneksi database dan API key, lalu jalankan:

```powershell
go run ./cmd/api
```

Hot reload tersedia melalui Air:

```powershell
air
```

## Tes dan pemeriksaan

```powershell
go test ./...
go build ./...
go vet ./...
```

Tes integrasi PostgreSQL opsional:

```powershell
$env:DAPODIK_TEST_DB = '1'
go test ./...
```

Tes integrasi membaca `.env` dan hanya menjalankan query baca.

## Deploy sebagai Windows Service

Build binary di Windows:

```powershell
go build -o dapodik-api.exe ./cmd/api
```

Salin `dapodik-api.exe`, `.env`, dan `install-service.ps1` ke folder yang sama di server Windows. Jalankan `install-service.ps1` dari PowerShell Administrator untuk memasang service otomatis. Lihat bagian Windows Service di [API_DOCS.md](API_DOCS.md).

Jangan commit `.env` atau binary hasil build. Gunakan `.env.example` sebagai template konfigurasi tanpa rahasia.
