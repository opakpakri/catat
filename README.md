# Catat — Cellix Aplikasi Transaksi & Arus Tunai

Aplikasi pencatat keuangan pribadi dan arus kas usaha mikro berbasis **Multi-Tenant** (*Row-Level Tenant Isolation* via `user_id`). Backend Go + PostgreSQL, frontend SvelteKit + Tailwind CSS.

```text
catat/
  backend/
    cmd/api/              entry point dan server HTTP Fiber
    internal/             config, middleware (JWT & Admin), models, handlers, routes
    migrations/           migrasi SQL (001_init_catat.sql)
  frontend/src/           aplikasi SvelteKit (transaksi, statistik, akun, pengguna, lainnya)
```

## Fitur
- **Transaksi (`/transaksi`):**
  - Ringkasan Saldo Bersih, Pendapatan, Pengeluaran, dan status kesehatan arus kas.
  - Empat tampilan terpadu: **Daily** (daftar harian), **Monthly** (kalender grid arus kas & acara), **Total** (ringkasan & progres anggaran), dan **Note** (daftar acara & *reminder* satu arah selesai).
  - Input cepat (`+ Catat Baru`): **Pendapatan**, **Pengeluaran**, **Transfer Antar Dompet**, dan **Acara / Reminder** dengan tanggal & waktu terpisah.
- **Statistik (`/statistik`):** Grafik pengeluaran per kategori, tren arus kas 6 bulan terakhir, serta manajemen anggaran (*Budgeting*) lengkap dengan indikator (`Aman`, `Mendekati Limit`, `Overlimit`) dan hapus anggaran.
- **Akun & Utang Piutang (`/akun`):** Multi-dompet (`CASH`, `BANK`, `EWALLET`, `INVESTMENT`), total kekayaan bersih, serta pencatatan **Utang & Piutang** yang terhubung langsung ke dompet (dengan validasi cicilan maksimal sesuai sisa tagihan dan pengingat otomatis di kalender).
- **Manajemen Pengguna (`/pengguna`):** Khusus `ADMIN` untuk membuat, mengedit, mencari, dan menghapus akun pengguna (`USER` / `ADMIN`); pendaftaran publik dinonaktifkan.
- **Lainnya (`/lainnya`):** Pilihan tema **Terang / Gelap / Sistem**, **Push Notifikasi Reminder Acara** (Service Worker + Notification API), manajemen kategori kustom, ekspor **CSV**, serta **Backup & Restore JSON**.
- **Panduan Interaktif:** Product tour per halaman menggunakan Driver.js dan navigasi responsif (Sidebar kiri di desktop, Bottom Navbar di mobile).

## Prasyarat
Go 1.24+, PostgreSQL 14+, Node.js 20+.

## Setup sekali jalan
1. Buat database: `CREATE DATABASE catat_db;`
2. Di `backend`: salin `.env.example` menjadi `.env` (lihat Konfigurasi di bawah) dan sesuaikan `DATABASE_URL`.
3. Di `backend`: `go mod tidy`. Di `frontend`: `npm install`.
4. Migrasi skema berjalan otomatis saat backend pertama kali dijalankan (`001_init_catat.sql`), tidak perlu menjalankan tool migrasi manual.

## Jalankan
Gunakan dua terminal:
- `cd backend` lalu `go run ./cmd/api` (`http://localhost:8080`)
- `cd frontend` lalu `npm run dev` (`http://localhost:5173`)

Saat pertama kali jalan, backend membuat akun Admin default jika belum ada:
- Email: `admin@catat.id`
- Password: `admin123`

Segera ganti password ini di lingkungan selain development.

## Konfigurasi
`backend/.env`:
```env
PORT=8080
DATABASE_URL=postgres://postgres:password@localhost:5432/catat_db?sslmode=disable
JWT_SECRET=ganti-dengan-secret-key-anda
CORS_ORIGINS=http://localhost:5173,http://127.0.0.1:5173
```
Untuk produksi, ganti `JWT_SECRET` dengan nilai acak yang kuat dan batasi `CORS_ORIGINS` ke domain frontend.

## Tests
- `cd backend` lalu `go test ./...` menjalankan unit test.
- Jika ada integration test, gunakan database sekali pakai (*throwaway*) terpisah, bukan database utama.

## Teknologi
- **Backend:** Go, Fiber v2, pgx v5, JWT, bcrypt
- **Frontend:** SvelteKit (Svelte 5), Tailwind CSS, Lucide Icons, Driver.js

## Produksi
Jalankan `npm run build` di `frontend` (menghasilkan static SPA di `frontend/build`), lalu sajikan lewat web server (mis. Caddy/Nginx) dan proxy `/api` ke backend Go. Pastikan `JWT_SECRET` dan `CORS_ORIGINS` sudah diatur untuk produksi.
