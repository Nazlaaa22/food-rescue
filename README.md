# Food Rescue API

Backend REST API untuk menyalurkan makanan berlebih yang masih layak konsumsi dari **Donatur**
(restoran, warung, katering, toko roti, acara kampus) ke **Mitra Penerima** (panti asuhan,
panti jompo, komunitas) melalui **Relawan**. Setiap perubahan status donasi tercatat supaya
penyaluran transparan.

Dibuat untuk UAS Pemrograman Backend Lanjut Praktikum, Kelompok 4,
Teknik Informatika, Fakultas Vokasi, Universitas Airlangga.

## Teknologi

| Teknologi          | Fungsi                                      |
| ------------------ | ------------------------------------------- |
| Go + Fiber v2      | REST API                                    |
| PostgreSQL         | Basis data                                  |
| JWT (HS256)        | Autentikasi                                 |
| bcrypt             | Hash password                               |
| Docker             | Menjalankan aplikasi di lingkungan konsisten |
| Clean Architecture | entity, repository, usecase, delivery       |

## Arsitektur

```
HTTP request
   │
   ▼
delivery/http  (route: baca request, panggil usecase, kirim response)
   │
   ▼
usecase        (aturan bisnis & validasi)
   │
   ▼
repository     (query SQL ke PostgreSQL)
   │
   ▼
entity         (struktur data & error domain, dipakai semua layer)
```

Ketergantungan hanya mengarah ke dalam: delivery → usecase → repository → entity.
Error dari usecase/repository memakai `entity.DomainError` lalu dipetakan ke status HTTP
di `delivery/http/helper.go`. Semua balasan memakai format yang sama dari `pkg/response.go`:

```json
{ "success": true, "message": "Daftar mitra penerima", "data": [ ... ] }
{ "success": false, "message": "mitra tidak ditemukan" }
```

## Struktur folder

```
food-rescue/
├── cmd/api/main.go            titik awal program
├── config/                    baca .env & koneksi database
├── internal/
│   ├── entity/                bentuk data & error domain
│   ├── repository/            akses database
│   ├── usecase/               logika bisnis
│   ├── delivery/http/         route/endpoint
│   └── middleware/            cek JWT, role, dan verifikasi akun
├── pkg/                       hash password, JWT, format response
├── migrations/                SQL pembuatan tabel (jalankan berurutan)
├── Dockerfile
└── .env.example
```

## Menjalankan secara lokal

1. **Clone repository**
   ```bash
   git clone <url-repo>
   cd food-rescue
   ```
2. **Buat database** `food_rescue` di PostgreSQL (lewat pgAdmin atau `createdb food_rescue`).
3. **Siapkan `.env`**: salin `.env.example` menjadi `.env`, lalu isi password PostgreSQL
   dan `JWT_SECRET` milikmu sendiri. File `.env` tidak boleh di-commit.
4. **Jalankan migration secara berurutan** (`001` sampai `008`), misalnya:
   ```bash
   for f in migrations/*.sql; do psql -U postgres -d food_rescue -f "$f"; done
   ```
5. **Jalankan server**
   ```bash
   go mod download
   go run ./cmd/api
   ```
   API berjalan di `http://localhost:8080` (atau sesuai `APP_PORT`).

### Membuat akun admin pertama

Registrasi publik hanya untuk `donor` dan `volunteer`. Untuk admin pertama:

1. Daftar lewat `POST /auth/register` seperti biasa (role `donor`).
2. Ubah role dan verifikasinya langsung di database:
   ```sql
   UPDATE users SET role = 'admin', is_verified = TRUE WHERE email = 'emailmu@contoh.com';
   ```
3. Login ulang supaya token baru membawa role `admin`.

Donatur dan relawan harus **diverifikasi admin** (`PUT /admin/users/:id/verify`) sebelum bisa
membuat donasi atau mengambil tugas antar.

### Menjalankan dengan Docker

```bash
docker build -t food-rescue .
docker run --env-file .env -p 8080:8080 food-rescue
```

Jika PostgreSQL berjalan di laptop (bukan di container), isi `DB_HOST=host.docker.internal` di `.env`.

## Alur status

```
Donasi:     available ──► allocated ──► picked_up ──► delivered
                 │            │
                 └────────────┴──► cancelled / expired

Distribusi: allocated ──► picked_up ──► delivered
                 └──► cancelled  (dibatalkan admin sebelum diambil relawan)

Pickup:     assigned ──► picked_up ──► delivered
                 └──► cancelled  (tugas kembali tersedia untuk relawan lain)
```

- Satu donasi boleh dibagi ke beberapa mitra. Total porsi alokasi tidak boleh melebihi porsi donasi.
- Donasi menjadi `delivered` setelah semua distribusinya selesai.
- Semua perubahan status donasi dicatat di tabel `donation_logs`.

## Daftar endpoint

Endpoint selain `/auth/*` membutuhkan header `Authorization: Bearer <token>`.

| Method | Endpoint                         | Role                     | Fungsi                                        |
| ------ | -------------------------------- | ------------------------ | --------------------------------------------- |
| POST   | `/auth/register`                 | Publik                   | Registrasi donatur/relawan                    |
| POST   | `/auth/login`                    | Publik                   | Login, mendapatkan token JWT                  |
| GET    | `/profile`                       | Semua role               | Lihat profil                                  |
| PUT    | `/profile`                       | Semua role               | Ubah nama & nomor telepon                     |
| GET    | `/categories`                    | Semua role               | Daftar kategori makanan                       |
| POST   | `/donations`                     | Donatur (terverifikasi)  | Buat donasi                                   |
| GET    | `/donations`                     | Donatur                  | Riwayat & status donasi sendiri               |
| GET    | `/pickups/available`             | Relawan (terverifikasi)  | Tugas antar yang tersedia                     |
| GET    | `/pickups/mine`                  | Relawan (terverifikasi)  | Riwayat tugas                                 |
| POST   | `/pickups`                       | Relawan (terverifikasi)  | Ambil tugas antar                             |
| PUT    | `/pickups/:id/status`            | Relawan (terverifikasi)  | Update status: picked_up/delivered/cancelled  |
| GET    | `/admin/users`                   | Admin                    | Daftar pengguna (`?role=`, `?is_verified=`)   |
| PUT    | `/admin/users/:id/verify`        | Admin                    | Verifikasi / batalkan verifikasi akun         |
| GET    | `/admin/categories`              | Admin                    | Daftar kategori                               |
| POST   | `/admin/categories`              | Admin                    | Tambah kategori                               |
| GET    | `/admin/donations`               | Admin                    | Semua donasi (`?status=`)                     |
| GET    | `/admin/partners`                | Admin                    | Daftar mitra (`?active=true`)                 |
| GET    | `/admin/partners/:id`            | Admin                    | Detail mitra                                  |
| POST   | `/admin/partners`                | Admin                    | Tambah mitra                                  |
| PUT    | `/admin/partners/:id`            | Admin                    | Ubah / nonaktifkan mitra                      |
| POST   | `/admin/distributions`           | Admin                    | Alokasikan porsi donasi ke mitra              |
| GET    | `/admin/distributions`           | Admin                    | Pantau distribusi (`?status=&donation_id=&partner_id=`) |
| GET    | `/admin/distributions/:id`       | Admin                    | Detail distribusi + status pickup             |
| PUT    | `/admin/distributions/:id/cancel`| Admin                    | Batalkan distribusi yang belum diambil        |
| GET    | `/admin/stats`                   | Admin                    | Statistik: porsi terselamatkan, dll.          |

### Contoh request

Alokasi 30 porsi dari donasi #1 ke mitra #2:

```http
POST /admin/distributions
Authorization: Bearer <token-admin>
Content-Type: application/json

{ "donation_id": 1, "partner_id": 2, "portions_allocated": 30 }
```

Contoh isi `data` dari `GET /admin/stats`:

```json
{
  "total_donations": 12,
  "donations_by_status": { "available": 3, "allocated": 2, "picked_up": 1, "delivered": 5, "expired": 0, "cancelled": 1 },
  "total_portions_donated": 640,
  "total_distributions": 15,
  "distributions_by_status": { "allocated": 3, "picked_up": 2, "delivered": 9, "cancelled": 1 },
  "portions_allocated": 520,
  "portions_rescued": 410,
  "rescue_rate_percent": 64.06,
  "total_partners": 6,
  "active_partners": 5,
  "total_donors": 8,
  "total_volunteers": 10,
  "unverified_users": 2
}
```

## Tim

| Nama                   | NIM       | Modul                                              |
| ---------------------- | --------- | -------------------------------------------------- |
| Nazlatul Khoiriyah     | 434241133 | Foundation, migration, autentikasi, donation logs  |
| Abdullah Azzam         | 434241118 | Kategori makanan, donasi, pickup, role middleware  |
| Nailurrachman Baldan B | 434241121 | Mitra, distribusi, statistik, format response      |
