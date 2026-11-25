# Spesifikasi Sistem: CATAT (Cellix Aplikasi Transaksi & Arus Tunai)

Dokumen ini berisi rancangan teknis, desain antarmuka (*UI/UX Design System*), arsitektur **Multi-Tenant**, sistem notifikasi latar belakang (**Web Push API & PWA**), skema database PostgreSQL, dan spesifikasi API untuk pengembangan aplikasi **CATAT (Cellix Aplikasi Transaksi & Arus Tunai)** menggunakan **Go (Golang)**, **Svelte**, **Tailwind CSS**, dan **PostgreSQL**.

## 1. Identitas & Tech Stack

* **Nama Aplikasi:** CATAT (*Cellix Aplikasi Transaksi & Arus Tunai*)

* **Segmen & Fokus:** Aplikasi pencatat keuangan pribadi dan arus kas usaha mikro yang mengintegrasikan pencatatan transaksi harian, kalender arus kas & pengingat acara (*event reminder* dengan notifikasi browser latar belakang), multi-dompet, penganggaran (*budgeting*), serta pencatatan utang/piutang dalam satu antarmuka terpadu (*mobile-first*).

* **Arsitektur Sistem:** Lightweight Multi-Tenant SaaS (Row-Level Tenant Isolation via `user_id`) — Ringan, cepat, dan terisolasi penuh antar pengguna.

* **Environment & Deployment (TANPA DOCKER):** **Dilarang menggunakan Docker (`Dockerfile` maupun `docker-compose.yml`).** Seluruh komponen aplikasi harus dijalankan secara *native* di sistem operasi menggunakan Go CLI, Node.js/Vite, dan koneksi langsung ke PostgreSQL lokal atau cloud database (seperti Neon PostgreSQL) melalui variabel `DATABASE_URL` pada file `.env`.

* **Backend:** Go (Golang) — Menggunakan *framework* **Fiber** atau **Echo** dengan JWT Authentication, PIN Security Verification, Tenant Isolation Middleware, Background Cron Scheduler (`time.NewTicker`), serta Web Push VAPID (`github.com/SherClockHolmes/webpush-go`).

* **Frontend:** SvelteKit / Svelte + Vite — SPA/SSR yang ringan, cepat, dan reaktif dengan tata letak *mobile-first* (Bottom Navigation Bar) serta dukungan **PWA (`manifest.json` + Service Worker `sw.js`)** untuk menerima notifikasi *push* saat aplikasi ditutup.

* **Styling:** Tailwind CSS — Utility-first CSS dengan dukungan *Dark Mode* (ala referensi *Money Manager*) maupun *Light Mode*.

* **Database:** PostgreSQL — Relational database dengan tipe data `NUMERIC(15, 2)` untuk presisi nominal uang serta indeks komposit `(user_id, ...)` untuk menjamin kecepatan agregasi kalender dan isolasi data.

* **Icons:** Lucide Icons (`lucide-svelte`) — Ikon garis bersih (*outline*) untuk navigasi utama, kategori transaksi, dan aksi cepat.

## 2. Arsitektur Multi-Tenant & Keamanan Isolasi Data

Setiap akun pada tabel `users` merepresentasikan satu **Tenant / Pengguna** yang independen. Untuk menjamin **User A tidak dapat melihat, mengakses, maupun mengubah data finansial milik User B dalam kondisi apa pun**, diterapkan 5 lapis pengamanan berikut:

1. **Foreign Key `user_id` Wajib di Semua Tabel:** Setiap tabel operasional (`wallets`, `categories`, `transactions`, `budgets`, `debts`, `calendar_events`, `push_subscriptions`) wajib memiliki kolom `user_id NOT NULL REFERENCES users(id) ON DELETE CASCADE`.

2. **JWT Context Extraction (Anti-Spoofing):** Backend **dilarang** menerima `user_id` dari payload JSON request atau query parameter URL. `user_id` harus diekstrak secara eksklusif dari *claims* token JWT yang sudah terverifikasi oleh middleware autentikasi.

3. **Strict Query Scoping:** Seluruh query database (`SELECT`, `UPDATE`, `DELETE`) wajib menyertakan klausa `WHERE user_id = $1`. Jika User A mencoba mengakses `/api/transactions/99` milik User B, sistem akan mengembalikan `404 Not Found` (bukan `403`, agar tidak membocorkan eksistensi ID milik tenant lain).

4. **Cross-Tenant Reference Validation & Atomic Balance Update:** Saat mencatat transaksi (`POST /api/transactions`), backend wajib memvalidasi kepemilikan `wallet_id` (dan `target_wallet_id` untuk transfer) milik `user_id`, lalu memperbarui saldo dompet dalam satu blok transaksi database (*ACID Transaction*).

5. **Keamanan Akses Tambahan (PIN Lock):** Mendukung proteksi 6-digit PIN (`security_pin_hash`) yang diverifikasi melalui endpoint khusus sebelum membuka layar utama aplikasi.

## 3. Arsitektur Web Push Notification (Background Reminder)

Agar fitur *reminder* acara pada kalender dapat mengirimkan notifikasi ke perangkat pengguna **meskipun tab browser/web CATAT sedang tidak dibuka**, sistem menggunakan standar **Web Push API + VAPID + Service Worker (PWA)** tanpa biaya pihak ketiga:

1. **Registrasi Service Worker & Subscription (Frontend Svelte):**
   * File `static/manifest.json` disediakan agar aplikasi dapat dipasang ke layar utama (*Add to Home Screen* — wajib untuk dukungan Web Push di Safari iOS).
   * File `static/sw.js` bertugas mendengarkan *event* `push` di latar belakang, menampilkan pop-up notifikasi sistem operasi (`self.registration.showNotification`), dan mengarahkan pengguna ke halaman kalender saat notifikasi diklik (`notificationclick`).
   * Saat pengguna mengaktifkan toggle *"Notifikasi Reminder"* (di menu `More` atau modal tambah acara), browser meminta izin (`Notification.requestPermission()`), mendaftarkan `PushSubscription` dengan `VAPID_PUBLIC_KEY`, dan mengirim `endpoint`, `p256dh`, serta `auth` ke backend Go (`POST /api/notifications/subscribe`).
2. **Background Scheduler & Pengiriman VAPID (Backend Go):**
   * Server Go menjalankan *goroutine* dengan `time.NewTicker(1 * time.Minute)` untuk memindai tabel `calendar_events` yang memiliki jadwal pengingat jatuh tempo (`remind_at <= NOW()` dan `is_notified = FALSE` dan `is_completed = FALSE`).
   * Untuk setiap acara yang jatuh tempo, backend mengambil semua perangkat terdaftar milik `user_id` dari tabel `push_subscriptions`, menandatangani *payload* JSON menggunakan `VAPID_PRIVATE_KEY` (`github.com/SherClockHolmes/webpush-go`), lalu mengirimkannya ke server *push* browser.
   * Jika server *push* mengembalikan status `410 Gone` atau `404 Not Found` (token perangkat sudah kedaluwarsa/dicabut pengguna), backend otomatis menghapus baris `push_subscriptions` tersebut agar database tetap bersih.

## 4. Design System & Palet Warna (Dark-First Financial UI)

Tampilan **CATAT** mengadopsi tata letak ringkas ala **Money Manager** dengan navigasi cepat berbasis periode bulanan (`< Nov 2026 >`), *Summary Bar* 3 kolom (*Income*, *Expenses*, *Total*), sub-tab terintegrasi (*Daily*, *Monthly*, *Total*), dan *Bottom Navigation Bar* dengan tombol aksi cepat (`+`) mengambang di pojok kanan bawah.

### Palet Warna (Tailwind CSS Mapping)

| Elemen UI | Warna / Efek | Kelas Tailwind CSS / Kode Hex | 
| ----- | ----- | ----- | 
| **Brand Logo & Accent** | Oranye Cerah | `text-[#EA580C]` (`orange-600`) | 
| **Background Utama** | Hitam Abu-abu Gelap | `bg-[#121212]` atau `bg-[#18181B]` (`zinc-950` / `zinc-900`) | 
| **Header, Card & Bottom Nav** | Abu-abu Gelap Elevated | `bg-[#1E1E22]` (`zinc-800`) | 
| **Garis Grid Kalender & Border** | Abu-abu Tipis | `border-[#2E2E34]` (`zinc-700/60`) | 
| **Primary Action & Tombol `+` (FAB)** | Oranye / Koral Energik | `bg-[#EA580C]` hover `bg-[#C2410C]` (`orange-600` / `orange-700`) | 
| **Active Sub-Tab Indicator** | Garis Bawah Oranye | `border-b-2 border-[#EA580C] text-white` | 
| **Nominal Pemasukan (*Income*)** | Biru Cerah | `text-[#3B82F6]` (`blue-500`) | 
| **Nominal Pengeluaran (*Expenses*)** | Merah Koral | `text-[#EF4444]` (`red-500`) | 
| **Nominal Total / Saldo Netral** | Putih Bersih | `text-[#F8FAFC]` (`slate-50`) | 
| **Badge Reminder / Acara Kalender** | Hijau Zamrud | `bg-[#10B981]/20 text-[#34D399]` (`emerald-500/20`) | 
| **Teks Utama** | Putih Lembut | `text-[#F4F4F5]` (`zinc-100`) | 
| **Teks Sekunder & Tanggal Grid** | Abu-abu Medium | `text-[#A1A1AA]` (`zinc-400`) | 

## 5. Struktur Navigasi & Fitur Utama (4 Menu Bottom Bar)

Antarmuka utama menggunakan **Top Period Header** (`< Nov 2026 >`, ikon *Search/Filter*, ikon *PIN Lock*) dan **Summary Bar** (*Income*, *Expenses*, *Total*) di bagian atas, serta **4 Menu Utama** di bagian bawah:

1. **Trans. (Transaksi & Kalender Terpadu)**

   * **Sub-Tab `Daily` (Daftar Harian):**

     * Menampilkan riwayat transaksi dalam bentuk daftar biasa (*list view*) yang dikelompokkan berdasarkan tanggal, lengkap dengan badge hari, total pemasukan harian (biru), dan total pengeluaran harian (merah).

   * **Sub-Tab `Monthly` (Kalender Grid & Reminder Acara):**

     * Menampilkan grid kalender bulanan (Minggu–Sabtu) seperti pada referensi desain.

     * Setiap sel tanggal menampilkan ringkasan angka **Pemasukan** (biru), **Pengeluaran** (merah), serta **Indikator/Badge Acara & Reminder** (misalnya: titik warna atau label singkat jadwal acara/jatuh tempo utang) secara langsung di dalam kotak tanggal.

     * *Interaksi Klik Tanggal:* Membuka *Bottom Sheet / Modal Detail Hari* yang menampilkan daftar transaksi dan daftar *reminder* acara pada tanggal tersebut, lengkap dengan tombol centang selesai (*toggle completed*) atau tambah baru.

   * **Sub-Tab `Total` (Ringkasan & Anggaran):**

     * Menampilkan ringkasan arus kas bulanan/tahunan serta progres anggaran (*Budgeting*) aktif.

   * **Tombol Aksi Cepat (`+` Floating Action Button):**

     * Satu tombol utama di pojok kanan bawah untuk membuka modal input cepat dengan pilihan tipe: **`Income`** (Pemasukan), **`Expense`** (Pengeluaran), **`Transfer`** (Pindah Saldo Antar Dompet), dan **`Reminder / Acara`** (Mencatat jadwal kegiatan + waktu pengiriman notifikasi *push* langsung ke kalender).

2. **Stats (Laporan & Anggaran)**

   * **Visualisasi Grafik:** *Donut/Pie Chart* untuk melihat persentase pengeluaran per kategori dan *Bar Chart* untuk tren arus kas bulanan.

   * **Anggaran (*Budgeting*):** Pengaturan batas maksimal pengeluaran per kategori per bulan lengkap dengan *progress bar* indikator visual (Aman, Mendekati Limit, Overlimit).

3. **Accounts (Multi-Dompet & Utang Piutang)**

   * **Tab Dompet (*Wallets*):** Menampilkan daftar sumber dana (Uang Tunai/Kas, Rekening Bank, E-Wallet), saldo masing-masing akun, serta akumulasi total kekayaan bersih.

   * **Tab Utang & Piutang (*Debts & Receivables*):** Pencatatan daftar pinjaman ke orang lain (*Piutang*) maupun pinjaman dari orang lain (*Utang*), progres cicilan, dan opsi otomatis membuat pengingat jatuh tempo di kalender `Monthly`.

4. **More (Kategori, Notifikasi, Ekspor & Keamanan)**

   * **Manajemen Kategori:** Tambah/edit kategori pemasukan dan pengeluaran kustom beserta pemilihan ikon dan warna.

   * **Pengaturan Notifikasi Push:** Aktivasi izin notifikasi browser (*Service Worker*), uji coba kirim notifikasi (*Test Push*), dan manajemen perangkat terdaftar.

   * **Ekspor Data & Backup:** Unduh laporan transaksi ke **CSV/Excel**, cetak laporan **PDF**, serta fitur **Backup & Restore (JSON)**.

   * **Keamanan Akses:** Pengaturan aktivasi dan perubahan **6-digit PIN** kunci aplikasi.

## 6. Skema Database Multi-Tenant (PostgreSQL)

Semua tabel operasional menggunakan `user_id` sebagai kunci isolasi tenant, dilengkapi indeks komposit untuk mempercepat agregasi kalender bulanan, pencarian riwayat transaksi, dan pemindaian *scheduler* notifikasi.

```sql
-- 1. Tabel Users (Setiap baris adalah 1 Tenant / Pengguna)
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    full_name VARCHAR(100) NOT NULL,
    email VARCHAR(150) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    security_pin_hash VARCHAR(255), -- Hash 6-digit PIN keamanan
    is_pin_enabled BOOLEAN DEFAULT FALSE,
    currency VARCHAR(10) NOT NULL DEFAULT 'IDR',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2. Tabel Dompet / Sumber Dana (Terisolasi per user_id)
CREATE TABLE wallets (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL, -- Contoh: 'Tunai', 'BCA', 'GoPay'
    type VARCHAR(20) NOT NULL DEFAULT 'CASH'
        CHECK (type IN ('CASH', 'BANK', 'EWALLET', 'INVESTMENT')),
    balance NUMERIC(15, 2) NOT NULL DEFAULT 0,
    color_hex VARCHAR(7) DEFAULT '#3B82F6',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, name)
);
CREATE INDEX idx_wallets_user_id ON wallets(user_id);

-- 3. Tabel Kategori Transaksi (Terisolasi per user_id)
CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    type VARCHAR(10) NOT NULL CHECK (type IN ('INCOME', 'EXPENSE')),
    icon VARCHAR(50) NOT NULL DEFAULT 'tag',
    color_hex VARCHAR(7) DEFAULT '#64748B',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, name, type)
);
CREATE INDEX idx_categories_user_type ON categories(user_id, type);

-- 4. Tabel Transaksi Keuangan (Untuk Tab Daily & Agregasi Kalender Monthly)
CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    wallet_id INT NOT NULL REFERENCES wallets(id) ON DELETE RESTRICT,
    target_wallet_id INT REFERENCES wallets(id) ON DELETE RESTRICT, -- Diisi jika type = 'TRANSFER'
    category_id INT REFERENCES categories(id) ON DELETE SET NULL,
    type VARCHAR(10) NOT NULL CHECK (type IN ('INCOME', 'EXPENSE', 'TRANSFER')),
    amount NUMERIC(15, 2) NOT NULL CHECK (amount > 0),
    transaction_date TIMESTAMP WITH TIME ZONE NOT NULL,
    note TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_transactions_user_date ON transactions(user_id, transaction_date DESC);

-- 5. Tabel Anggaran / Budgeting (Terisolasi per user_id)
CREATE TABLE budgets (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id INT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    amount_limit NUMERIC(15, 2) NOT NULL CHECK (amount_limit > 0),
    period_month INT NOT NULL CHECK (period_month BETWEEN 1 AND 12),
    period_year INT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, category_id, period_month, period_year)
);
CREATE INDEX idx_budgets_user_period ON budgets(user_id, period_year, period_month);

-- 6. Tabel Catatan Utang & Piutang (Terisolasi per user_id)
CREATE TABLE debts (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    person_name VARCHAR(100) NOT NULL,
    type VARCHAR(15) NOT NULL CHECK (type IN ('DEBT', 'RECEIVABLE')), -- DEBT: Utang, RECEIVABLE: Piutang
    total_amount NUMERIC(15, 2) NOT NULL CHECK (total_amount > 0),
    paid_amount NUMERIC(15, 2) NOT NULL DEFAULT 0,
    due_date DATE,
    status VARCHAR(15) NOT NULL DEFAULT 'UNPAID'
        CHECK (status IN ('UNPAID', 'PARTIAL', 'PAID')),
    note TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_debts_user_status ON debts(user_id, status);

-- 7. Tabel Kalender & Reminder Acara (Terintegrasi di Tab Monthly & Web Push)
CREATE TABLE calendar_events (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    debt_id INT REFERENCES debts(id) ON DELETE SET NULL, -- Terhubung otomatis jika dari jatuh tempo utang
    title VARCHAR(150) NOT NULL,
    description TEXT,
    event_date DATE NOT NULL, -- Tanggal tampil pada grid kalender Monthly
    event_time TIME, -- Opsional: jam acara
    remind_at TIMESTAMP WITH TIME ZONE, -- Waktu spesifik pengiriman Web Push Notification
    is_notified BOOLEAN DEFAULT FALSE, -- Penanda apakah notifikasi push sudah terkirim
    is_completed BOOLEAN DEFAULT FALSE,
    color_hex VARCHAR(7) DEFAULT '#10B981',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_calendar_events_user_date ON calendar_events(user_id, event_date);
CREATE INDEX idx_calendar_events_pending_push ON calendar_events(is_notified, is_completed, remind_at);

-- 8. Tabel Langganan Perangkat Web Push Notification (Terisolasi per user_id)
CREATE TABLE push_subscriptions (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    endpoint TEXT NOT NULL UNIQUE,
    p256dh TEXT NOT NULL,
    auth TEXT NOT NULL,
    user_agent VARCHAR(255), -- Info browser/perangkat (misal: Chrome Android, Edge Windows)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_push_subscriptions_user_id ON push_subscriptions(user_id);
```

## 7. Rancangan Endpoint REST API (Backend Go)

Seluruh rute di bawah `/api/*` (kecuali `/api/auth/*`) dilindungi oleh **JWT & Tenant Context Middleware** yang menyuntikkan `userID` ke `c.Locals("userID")`.

| Modul | Method | Endpoint | Deskripsi & Isolasi Tenant | 
| ----- | ----- | ----- | ----- | 
| **Auth & Security** | `POST` | `/api/auth/register` | Registrasi pengguna baru + auto-seed dompet & kategori default | 
|  | `POST` | `/api/auth/login` | Login pengguna & menerbitkan token JWT berisi `user_id` | 
|  | `POST` | `/api/auth/pin/setup` | Mengatur atau mengubah 6-digit PIN keamanan (`WHERE id = $1`) | 
|  | `POST` | `/api/auth/pin/verify` | Memverifikasi 6-digit PIN untuk membuka kunci aplikasi | 
| **Web Push Notif** | `GET` | `/api/notifications/vapid-public-key` | Mengambil `VAPID_PUBLIC_KEY` untuk registrasi Push API di browser | 
|  | `POST` | `/api/notifications/subscribe` | Menyimpan `PushSubscription` (`endpoint`, `p256dh`, `auth`) untuk `user_id` | 
|  | `DELETE` | `/api/notifications/unsubscribe` | Menghapus langganan notifikasi perangkat tertentu (`WHERE user_id = $1 AND endpoint = $2`) | 
|  | `POST` | `/api/notifications/test` | Mengirim notifikasi uji coba langsung ke perangkat milik `user_id` | 
| **Trans. (Daily)** | `GET` | `/api/transactions` | List transaksi harian milik `user_id` (`?month=&year=&wallet_id=&category_id=&search=`) | 
|  | `POST` | `/api/transactions` | Catat transaksi baru (`INCOME`/`EXPENSE`/`TRANSFER`) & update saldo dompet secara atomik | 
|  | `PUT` | `/api/transactions/:id` | Edit transaksi & sesuaikan ulang selisih saldo dompet (`WHERE id = $1 AND user_id = $2`) | 
|  | `DELETE` | `/api/transactions/:id` | Hapus transaksi & kembalikan saldo dompet (`WHERE id = $1 AND user_id = $2`) | 
| **Trans. (Monthly)** | `GET` | `/api/calendar/monthly` | **Endpoint Gabungan:** Mengembalikan agregasi pemasukan/pengeluaran per tanggal + daftar *reminder* acara dalam 1 bulan (`?month=&year=`) | 
|  | `POST` | `/api/events` | Menambahkan *reminder* acara baru beserta `remind_at` ke kalender untuk `user_id` | 
|  | `PUT` | `/api/events/:id` | Edit jadwal, detail acara, atau waktu `remind_at` (`WHERE id = $1 AND user_id = $2`) | 
|  | `PATCH` | `/api/events/:id/toggle` | Menandai acara/reminder selesai atau belum (`WHERE id = $1 AND user_id = $2`) | 
|  | `DELETE` | `/api/events/:id` | Hapus acara dari kalender (`WHERE id = $1 AND user_id = $2`) | 
| **Stats & Budgets** | `GET` | `/api/reports/summary` | Ringkasan arus kas & data grafik pengeluaran per kategori (`?month=&year=`) | 
|  | `GET` | `/api/budgets` | List anggaran bulanan beserta total realisasi pengeluaran (`?month=&year=`) | 
|  | `POST` | `/api/budgets` | Set atau update batas anggaran kategori (`UPSERT` berdasarkan `user_id`) | 
| **Accounts & Debts** | `GET` | `/api/wallets` | List dompet & total saldo milik `user_id` | 
|  | `POST` | `/api/wallets` | Tambah dompet / sumber dana baru untuk `user_id` | 
|  | `PUT` | `/api/wallets/:id` | Edit nama/warna dompet (`WHERE id = $1 AND user_id = $2`) | 
|  | `GET` | `/api/debts` | List catatan utang & piutang milik `user_id` (`?type=&status=`) | 
|  | `POST` | `/api/debts` | Catat utang/piutang baru (opsional auto-insert ke `calendar_events` beserta `remind_at`) | 
|  | `PATCH` | `/api/debts/:id/pay` | Mencatat pembayaran cicilan atau pelunasan utang/piutang | 
| **More (Kategori & Ekspor)** | `GET` / `POST` | `/api/categories` | Mengambil daftar atau menambah kategori kustom milik `user_id` | 
|  | `DELETE` | `/api/categories/:id` | Hapus kategori kustom (`WHERE id = $1 AND user_id = $2`) | 
|  | `GET` | `/api/export/csv` | Unduh riwayat transaksi pengguna dalam format `.csv` | 
|  | `GET` / `POST` | `/api/backup/json` | Unduh (*export*) atau pulihkan (*restore*) data JSON milik `user_id` | 

## 8. Struktur Folder & Cara Menjalankan (Native Tanpa Docker)

**Penting untuk AI Agent / Developer:** Jangan membuat file `docker-compose.yml` atau `Dockerfile`. Jalankan proyek secara *native* di host/lokal dengan struktur berikut:

```text
catat/
├── backend/                  # Golang (Fiber + pgx/GORM + webpush-go) - Native Binary
│   ├── cmd/api/main.go       # Entry point aplikasi CATAT API & Background Reminder Worker
│   ├── internal/
│   │   ├── config/           # Koneksi PostgreSQL via .env, JWT_SECRET, & VAPID Keys
│   │   ├── middleware/       # auth_tenant.go (Validasi JWT & Ekstrak user_id)
│   │   ├── models/           # Struct Go & Request/Response DTO
│   │   ├── handlers/         # HTTP Handlers dengan strict WHERE user_id = ?
│   │   ├── workers/          # reminder_scheduler.go (Ticker 1 menit kirim Web Push)
│   │   └── routes/           # Definisi endpoint API
│   ├── migrations/           # 001_init_catat.sql
│   ├── .env                  # DATABASE_URL, JWT_SECRET, VAPID_PUBLIC_KEY, VAPID_PRIVATE_KEY
│   └── go.mod
│
└── frontend/                 # Svelte + Tailwind CSS + PWA Service Worker
    ├── static/
    │   ├── manifest.json     # Konfigurasi PWA (Installable di Android/iOS/Desktop)
    │   ├── sw.js             # Service Worker untuk menangkap event 'push' & 'notificationclick'
    │   └── icons/            # Ikon PWA (192x192 & 512x512)
    ├── src/
    │   ├── lib/
    │   │   ├── api.js        # Fetch wrapper otomatis menyisipkan Bearer Token
    │   │   ├── pushClient.js # Helper registrasi Service Worker & PushSubscription
    │   │   ├── stores/       # authStore.js, periodStore.js (Bulan/Tahun aktif), lockStore.js
    │   │   ├── utils/        # formatCurrency.js (IDR), calendarGrid.js (Helper grid bulanan)
    │   │   └── components/
    │   │       ├── layout/   # TopHeader.svelte, SummaryBar.svelte, BottomNav.svelte
    │   │       ├── views/    # DailyList.svelte, MonthlyCalendar.svelte, TotalSummary.svelte
    │   │       ├── modals/   # QuickAddModal.svelte (Transaksi & Event), DayDetailModal.svelte
    │   │       └── security/ # PinLockOverlay.svelte
    │   ├── routes/           # Login, Register, Trans (Daily/Monthly), Stats, Accounts, More
    │   └── app.css           # Konfigurasi Tailwind CSS
    ├── tailwind.config.js
    └── package.json
```

### Perintah Eksekusi Lokal (Native)

1. **Database:** Jalankan skema `migrations/001_init_catat.sql` langsung ke PostgreSQL lokal atau cloud database (Neon) menggunakan `psql` atau SQL client, lalu isi `DATABASE_URL` pada `backend/.env`.
2. **Generate VAPID Keys & Backend Go:** Masuk ke folder `backend/`, jalankan `go mod tidy`, isi `VAPID_PUBLIC_KEY` dan `VAPID_PRIVATE_KEY` di `.env`, lalu jalankan `go run cmd/api/main.go`.
3. **Frontend Svelte:** Masuk ke folder `frontend/`, jalankan `npm install` dan `npm run dev` (akses melalui `http://localhost:5173` agar Service Worker aktif).