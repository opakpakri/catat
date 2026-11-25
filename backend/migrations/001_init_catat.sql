-- 1. Tabel Users (Setiap baris adalah 1 Tenant / Pengguna)
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    full_name VARCHAR(100) NOT NULL,
    email VARCHAR(150) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'USER' CHECK (role IN ('ADMIN', 'USER')),
    security_pin_hash VARCHAR(255), -- Hash 6-digit PIN keamanan
    is_pin_enabled BOOLEAN DEFAULT FALSE,
    currency VARCHAR(10) NOT NULL DEFAULT 'IDR',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(20) NOT NULL DEFAULT 'USER';

-- Default Admin Account (Email: admin@catat.id | Password: admin123)
INSERT INTO users (full_name, email, password_hash, role, currency)
VALUES (
    'Administrator CATAT',
    'admin@catat.id',
    '$2a$10$Bi09JJ9eTusWI6EgrdumxOpQTWfLNbmKsowxqOrggQLfIBvhy6.CW',
    'ADMIN',
    'IDR'
)
ON CONFLICT (email) DO UPDATE SET role = 'ADMIN';

-- 2. Tabel Dompet / Sumber Dana (Terisolasi per user_id)
CREATE TABLE IF NOT EXISTS wallets (
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
CREATE INDEX IF NOT EXISTS idx_wallets_user_id ON wallets(user_id);

-- 3. Tabel Kategori Transaksi (Terisolasi per user_id)
CREATE TABLE IF NOT EXISTS categories (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    type VARCHAR(10) NOT NULL CHECK (type IN ('INCOME', 'EXPENSE')),
    icon VARCHAR(50) NOT NULL DEFAULT 'tag',
    color_hex VARCHAR(7) DEFAULT '#64748B',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, name, type)
);
CREATE INDEX IF NOT EXISTS idx_categories_user_type ON categories(user_id, type);

-- 4. Tabel Transaksi Keuangan (Untuk Tab Daily & Agregasi Kalender Monthly)
CREATE TABLE IF NOT EXISTS transactions (
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
CREATE INDEX IF NOT EXISTS idx_transactions_user_date ON transactions(user_id, transaction_date DESC);

-- 5. Tabel Anggaran / Budgeting (Terisolasi per user_id)
CREATE TABLE IF NOT EXISTS budgets (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id INT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    amount_limit NUMERIC(15, 2) NOT NULL CHECK (amount_limit > 0),
    period_month INT NOT NULL CHECK (period_month BETWEEN 1 AND 12),
    period_year INT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, category_id, period_month, period_year)
);
CREATE INDEX IF NOT EXISTS idx_budgets_user_period ON budgets(user_id, period_year, period_month);

-- 6. Tabel Catatan Utang & Piutang (Terisolasi per user_id)
CREATE TABLE IF NOT EXISTS debts (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    wallet_id INT REFERENCES wallets(id) ON DELETE SET NULL,
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
ALTER TABLE debts ADD COLUMN IF NOT EXISTS wallet_id INT REFERENCES wallets(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_debts_user_status ON debts(user_id, status);

-- 7. Tabel Kalender & Reminder Acara (Terintegrasi di Tab Monthly)
CREATE TABLE IF NOT EXISTS calendar_events (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    debt_id INT REFERENCES debts(id) ON DELETE SET NULL, -- Terhubung otomatis jika dari jatuh tempo utang
    title VARCHAR(150) NOT NULL,
    description TEXT,
    event_date DATE NOT NULL, -- Tanggal tampil pada grid kalender Monthly
    event_time TIME, -- Opsional: jam acara
    is_completed BOOLEAN DEFAULT FALSE,
    color_hex VARCHAR(7) DEFAULT '#10B981',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_calendar_events_user_date ON calendar_events(user_id, event_date);
