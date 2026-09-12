# Location Analysis API

Backend Go untuk menganalisis beberapa layer PostGIS berdasarkan latitude dan longitude.

## Alur request

1. Route menerima request HTTP.
2. Middleware memasang timeout, rate limit, API key, recovery, dan security headers.
3. Controller memvalidasi `lat` dan `lng`.
4. Service memeriksa Redis menggunakan koordinat sebagai cache key.
5. Jika cache tidak tersedia, satu query gabungan dijalankan ke PostgreSQL.
6. Hasil database diberi label dan warna oleh backend, lalu disimpan ke Redis.
7. Controller mengembalikan JSON kepada client.

## Menjalankan aplikasi

1. Siapkan `.env`.
2. Isi `DATABASE_URL`, `API_KEY`, dan `REDIS_URL` jika cache digunakan.
3. Jalankan `go run .`.
4. Panggil endpoint dengan header `X-API-Key`:

```bash
curl "http://localhost:3101/api/analyze?lat=-8.65&lng=115.2167" \
  -H "X-API-Key: API_KEY_ANDA"
```

Redis bersifat opsional. Jika `REDIS_URL` kosong atau Redis sedang tidak tersedia, query tetap dijalankan langsung ke PostgreSQL.

## Menjalankan dengan Docker Compose

Compose menjalankan dua container:

- `api`: backend Go pada port internal yang ditentukan oleh `PORT`.
- `redis`: cache internal yang hanya dapat diakses oleh container API.

File `.env` tetap disimpan di host VPS, bersebelahan dengan `compose.yaml`. Nilainya diinjeksi ketika container dibuat dan tidak disalin ke image Docker.

Siapkan environment:

```bash
copy env dari mac ke vps
```

Minimal isi nilai berikut:

```dotenv
APP_PORT=3100
PORT=3101
DATABASE_URL=postgresql://username:password@hostname:5432/database_name?sslmode=require
REDIS_URL=redis://redis:6379/0
API_KEY=ganti-dengan-random-string-minimal-32-karakter
GIN_MODE=release
ENABLE_HSTS=false
TRUSTED_PROXIES=
RATE_LIMIT_PER_MINUTE=60
REQUEST_TIMEOUT=5s
ANALYZE_CACHE_TTL=10m
ANALYZE_CACHE_VERSION=v1
TZ=Asia/Makassar
```

Compose meneruskan seluruh variabel dari `.env` ke container API melalui `env_file`. Untuk Redis, hostname `redis` mengacu pada service Redis di jaringan internal Docker.

Validasi konfigurasi lalu jalankan:

```bash
docker compose config
docker compose up -d --build
docker compose ps
```

Periksa health dan log:

```bash
curl http://localhost:3100/health
curl http://localhost:3100/ready
docker compose logs -f api redis
```

Memanggil endpoint analyze:

```bash
curl "http://localhost:3100/api/analyze?lat=-8.65&lng=115.2167" \
  -H "X-API-Key: API_KEY_ANDA"
```

Update aplikasi setelah source code berubah:

```bash
git pull
docker compose up -d --build
```

Menghentikan stack:

```bash
docker compose down
```

Redis tidak mempublikasikan port `6379` ke internet dan digunakan sebagai cache tanpa persistence. Container Redis diberi batas cache `128mb` dengan eviction policy `allkeys-lru`.

`ANALYZE_CACHE_TTL` menentukan kapan cache terhapus otomatis. Setelah isi tabel spasial diperbarui, naikkan `ANALYZE_CACHE_VERSION`, misalnya dari `v1` menjadi `v2`, lalu buat ulang container. Cara ini membuat cache lama langsung diabaikan tanpa menjalankan penghapusan Redis secara menyeluruh.

## Menambah endpoint

1. Buat controller baru di `controllers/`.
2. Buat service baru di `services/` jika endpoint mempunyai logika atau akses data.
3. Tambahkan constructor controller tersebut di `main.go`.
4. Tambahkan controller ke struct `routes.Controllers`, lalu daftarkan endpoint di `routes/routes.go`.

Semua endpoint di dalam grup `/api` otomatis memakai API key dan rate limit yang dipasang di `main.go`.

## Dokumentasi Postman

Import file berikut ke Postman:

- `docs/postman/location-analysis.postman_collection.json`
- `docs/postman/development.postman_environment.json`
- `docs/postman/production.postman_environment.json`

Isi variable `api_key` langsung melalui Postman dan jangan menyimpan API key asli ke repository. Collection berisi endpoint health, readiness, analyze, parameter koordinat, serta contoh respons status `200`, `400`, `401`, `404`, `429`, `500`, dan `503`.

## Mengubah pencarian layer

Konfigurasi layer berada di `services/analyze_layers.go`. Setiap layer memilih salah satu tipe pencarian berikut:

```go
SearchIntersect
SearchRadius
SearchGrouped
```

Contoh mengubah flood dari intersect menjadi radius:

```go
{
    Name:         "flood",
    Table:        "flood_zones",
    ValueColumns: []string{"area"},
    SearchType:   SearchRadius,
    Attributes:   []string{"area"},
    Limit:        5,
},
```

Query gabungan dibuat satu kali ketika `NewAnalyzeService` dijalankan. Konfigurasi yang tidak valid membuat aplikasi berhenti saat startup, bukan saat menerima request. Nama tabel, kolom, dan atribut hanya boleh berasal dari konfigurasi internal ini; jangan mengisinya dari query parameter pengguna.

Perubahan konfigurasi query otomatis menghasilkan namespace cache Redis yang baru. Karena itu respons dengan strategi pencarian lama tidak akan terbaca setelah aplikasi di-deploy ulang.

## Menambah layer analisis

Penambahan layer tidak memerlukan perubahan pada controller atau route `/api/analyze`:

1. Tambahkan satu `LayerDefinition` di `services/analyze_layers.go`.
2. Pilih `SearchIntersect`, `SearchRadius`, atau `SearchGrouped`.
3. Daftarkan hanya atribut yang boleh muncul dalam respons API.
4. Tambahkan warna dan label di `services/analyze_style.go` jika layer mempunyai style khusus.
5. Jalankan pemeriksaan format, `go vet`, dan `go build ./...` sebelum deployment.

`config/config.go` hanya menyimpan konfigurasi runtime seperti database, Redis, API key, proxy, dan rate limit. Detail tabel serta style tetap berada dekat dengan fitur analyze.

Kolom internal `style_value` dipakai service untuk menentukan `label` dan `color`, tetapi tidak dikirim ke pengguna. Bentuk setiap hasil adalah:

```json
{
  "layer": "temperature",
  "distance_meters": 0,
  "label": "Hot",
  "color": "#FF9500",
  "attributes": {
    "suhu": "Hot"
  }
}
```

## Keamanan Backend

### Sudah diterapkan

- SQL Injection dicegah dengan parameterized query serta whitelist nama tabel dan kolom.
- Koordinat divalidasi sebagai angka dengan batas latitude dan longitude yang benar.
- Endpoint bisnis dilindungi menggunakan API key melalui header `X-API-Key`.
- API key dibandingkan secara constant-time untuk mengurangi risiko timing attack.
- Credential database, Redis, dan API key dibaca dari environment dan tidak di-hardcode dalam source code.
- Request API memiliki timeout agar query yang terlalu lama dapat dibatalkan.
- Rate limiter per IP membatasi jumlah request yang diterima setiap menit.
- Security headers mencegah MIME sniffing, framing, pengiriman referrer, dan penyimpanan respons pada cache client.
- HSTS dapat diaktifkan melalui environment setelah backend menggunakan HTTPS.
- Trusted proxy dikonfigurasi agar alamat IP client tidak dipercaya dari proxy sembarangan.
- Panic recovery mencegah satu request bermasalah mematikan seluruh server.
- Pesan error untuk client dibuat umum sedangkan detail kesalahan hanya dicatat pada log server.
- Container API berjalan sebagai user non-root dengan filesystem read-only dan `no-new-privileges`.
- Redis hanya tersedia melalui jaringan internal Docker dan tidak mempublikasikan port ke internet.

### Dapat dikembangkan berikutnya

- Tambahkan login dan authentication agar setiap request dapat dihubungkan dengan user tertentu.
- Tambahkan rate limiter per user setelah authentication tersedia tanpa menghapus perlindungan per IP.
- Pindahkan penyimpanan rate limit ke Redis ketika backend dijalankan pada lebih dari satu instance.
- Tambahkan rotasi dan pencabutan API key agar credential yang bocor dapat segera dinonaktifkan.
- Tambahkan role-based access control jika setiap user mempunyai hak akses endpoint yang berbeda.
- Gunakan secret manager atau Docker Secrets untuk melindungi credential production dengan lebih baik.
- Gunakan HTTPS melalui reverse proxy dan aktifkan HSTS hanya setelah sertifikat berfungsi dengan benar.
- Tambahkan request ID, audit log, monitoring, dan alert agar aktivitas mencurigakan lebih mudah ditelusuri.
- Tambahkan aturan CORS berbasis allowlist jika backend nantinya diakses oleh frontend web dari browser.
- Tambahkan automated test, vulnerability scanning, dan dependency scanning sebelum deployment production.
- Gunakan user PostgreSQL khusus dengan hak akses minimum sesuai tabel yang dibutuhkan backend.

## CI/CD

Workflow `.github/workflows/ci.yml` menjalankan CI berupa pemeriksaan format, `go vet`, build Go, dan build Docker pada setiap push serta pull request. Deployment ke VPS masih dilakukan secara manual dengan langkah berikut.

### Deployment pertama di VPS

1. Pastikan Git, GitHub CLI, Docker, dan Docker Compose sudah terpasang.
2. Login ke GitHub dari VPS:

```bash
gh auth login
```

Saat diminta, pilih GitHub.com, protokol HTTPS, lalu ikuti proses login melalui browser. Periksa hasil login:

```bash
gh auth status
```

3. Clone repository. Ganti `PEMILIK_REPOSITORY/NAMA_REPOSITORY` dengan repository Anda:

```bash
gh repo clone PEMILIK_REPOSITORY/NAMA_REPOSITORY ~/expressXgolang
cd ~/expressXgolang/golang
```

4. Buat environment production:

```bash
nano .env
chmod 600 .env
```

Isi `DATABASE_URL`, `API_KEY`, port, dan konfigurasi production lainnya. Jangan commit file `.env`.

5. Validasi lalu jalankan aplikasi:

```bash
docker compose config --quiet
docker compose up -d --build
docker compose ps
```

6. Periksa API dan log:

```bash
curl --fail http://127.0.0.1:3100/health
curl --fail http://127.0.0.1:3100/ready
docker compose logs --tail 100 api redis
```

### Update setelah ada commit baru

Masuk ke folder project di VPS:

```bash
cd ~/expressXgolang/golang
```

Ambil commit terbaru dari branch `main`:

```bash
git pull --ff-only origin main
```

Build ulang dan jalankan container:

```bash
docker compose up -d --build
docker compose ps
```

Periksa deployment:

```bash
curl --fail http://127.0.0.1:3100/health
curl --fail http://127.0.0.1:3100/ready
docker compose logs --tail 100 api redis
```

File `.env` tetap berada di VPS karena masuk `.gitignore`, sehingga `git pull` tidak mengganti credential production.
