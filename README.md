# BillingRevaktor (Billing ISP Management System)

Sistem Billing ISP terintegrasi adalah platform manajemen pelanggan, langganan, dan penagihan internet yang dikembangkan menggunakan **Go Backend (Clean Architecture)** dan **Vue 3 Frontend (Vite + TypeScript + Vuetify)**. Platform ini dirancang untuk menyederhanakan alur bisnis ISP mulai dari aktivasi langganan, pembuatan invoice otomatis via payment gateway (Xendit), penanganan isolasi otomatis (suspend/unsuspend) pada MikroTik, hingga integrasi monitoring perangkat jaringan (OLT & ODP).

---

## 1. Tech Stack (Teknologi yang Digunakan)

### **Backend**
- **Language**: Go (Golang) v1.25.6+
- **HTTP Web Framework**: Gin-Gonic
- **ORM**: GORM (Object Relational Mapping)
- **Database**: MySQL
- **Scheduler**: `github.com/robfig/cron/v3` (Diintegrasikan dengan manajemen dinamis berbasis DB)
- **Encryption**: Fernet Encryption Service (Untuk mengamankan data rahasia seperti konfigurasi server)
- **JWT**: JSON Web Tokens untuk autentikasi berbasis role-permission

### **Frontend**
- **Framework**: Vue 3 (Composition API)
- **Build Tool**: Vite + TypeScript
- **UI Framework**: Vuetify 3 (Material Design Components & Icons)
- **State Management**: Pinia
- **Routing**: Vue Router
- **HTTP Client**: Axios

---

## 2. Fitur Utama

- **Dashboard Real-time & Analitik**: Menampilkan ringkasan statistik pendapatan harian/bulanan, jumlah tiket aktif, status jaringan, dan pelanggan bermasalah.
- **Manajemen Pelanggan & Langganan**: Pengelolaan daftaran pelanggan, status modem/layanan, kalkulasi biaya prorate, diskon dinamis, dan pencatatan history.
- **Pembuatan Invoice Otomatis & Manual**:
  - Secara terjadwal men-generate invoice tagihan H-5 sebelum jatuh tempo secara otomatis.
  - Opsi pembuatan invoice manual langsung terintegrasi dengan Xendit Payment Link.
- **Webhook Payment Callback**: Integrasi callback instan dengan Xendit untuk memproses pelunasan invoice dan memicu pembukaan isolasi secara otomatis.
- **Auto-Suspend & Unsuspend (MikroTik RouterOS)**:
  - Tugas terjadwal otomatis mengisolasi pelanggan menunggak (jatuh tempo melewati batas bayar) di MikroTik PPPoE/Hotspot.
  - Setelah invoice dibayar via Xendit, sistem langsung memanggil API MikroTik untuk me-reactivate (unsuspend) layanan pelanggan tanpa intervensi manual.
- **Dynamic Scheduler Control Panel**:
  - Manajemen penjadwalan tugas latar belakang (Cronjob) secara langsung melalui Web UI (`Pengaturan Sistem`).
  - Mengubah jadwal tugas (cron expression) dan status aktif/nonaktif di memori tanpa perlu restart server.
  - Tombol **"Jalankan Sekarang"** untuk memicu eksekusi *asynchronous* manual secara instan dengan indikator status real-time (*Running/Success/Dipicu*).
- **Trouble Ticket (Pelaporan Keluhan)**: Tiket penanganan keluhan teknis pelanggan terintegrasi dengan penugasan teknisi dan log aktivitas.
- **Manajemen Inventaris & Topology Perangkat**: Visualisasi topologi jaringan OLT ke ODP dan pendataan barang logistik.

---

## 3. Struktur Direktori Proyek

```plaintext
BillingRevaktor/
├── backend/                  # Kode Backend Golang
│   ├── cmd/api/              # Entry point aplikasi (main.go)
│   ├── config/               # Load configuration env
│   ├── internal/             # Logika internal (Clean Architecture)
│   │   ├── delivery/http/    # HTTP Handlers & Routes (REST API)
│   │   ├── domain/           # Entities & Interface Kontrak
│   │   ├── repository/       # Data Access / Database Query (GORM)
│   │   ├── scheduler/        # Dynamic Cron Scheduler Manager
│   │   └── usecase/          # Business Logic Layer
│   ├── pkg/                  # Library/Utility Pendukung (Database, Logger, Utils)
│   └── uploads/              # Folder penyimpanan file unggahan
├── frontend/                 # Kode Frontend Vue 3
│   ├── src/
│   │   ├── assets/           # Gambar, static files, css
│   │   ├── components/       # Komponen reusable & dialog
│   │   ├── layouts/          # DefaultLayout & AuthLayout
│   │   ├── services/         # Integrasi API (axios client)
│   │   ├── stores/           # Pinia State Management
│   │   ├── views/            # Halaman Dashboard, Pelanggan, Settings, dll.
│   │   └── router/           # Konfigurasi Navigasi & Guarding
│   ├── package.json          # File package-depedency node
│   └── vite.config.ts        # Konfigurasi Vite bundler
├── migrations/               # File SQL skema basis data DDL
├── postman/                  # File koleksi API test Postman
└── README.md                 # Dokumentasi ini
```

---

## 4. Cara Menjalankan Aplikasi

### **A. Pengaturan Database**
1. Pastikan server MySQL Anda aktif.
2. Buat database kosong bernama `billing_revaktor`.
3. Skema tabel database akan di-migrate secara otomatis oleh GORM AutoMigrate saat pertama kali Anda menyalakan backend Golang.

### **B. Menjalankan Backend (Golang)**
1. Masuk ke direktori backend:
   ```bash
   cd backend
   ```
2. Duplikat file `.env.example` menjadi `.env`:
   ```bash
   cp .env.example .env
   ```
3. Buka file `.env` dan sesuaikan kredensial Anda, seperti database URL, API Key Xendit, token callback, Mapbox key, dan Telegram bot token.
4. Jalankan backend menggunakan perintah:
   ```bash
   go run ./cmd/api
   ```
5. Server backend akan berjalan secara default pada port `8000` (`http://localhost:8000`).

### **C. Menjalankan Frontend (Vue 3)**
1. Masuk ke direktori frontend:
   ```bash
   cd frontend
   ```
2. Duplikat file `.env.example` menjadi `.env.local` atau `.env`:
   ```bash
   cp .env.example .env.local
   ```
3. Sesuaikan alamat base URL API pada file tersebut jika backend Anda tidak berjalan pada port standard (default: `http://localhost:8000`).
4. Instal paket dependensi:
   ```bash
   npm install
   ```
5. Jalankan frontend dengan mode developer:
   ```bash
   npm run dev
   ```
6. Aplikasi frontend akan dapat diakses melalui browser Anda pada URL default `http://localhost:3000`.

---

## 5. Konfigurasi Jadwal Tugas Latar Belakang (Dynamic Scheduler)

Melalui menu **Pengaturan Sistem** di antarmuka Web Admin, Anda dapat mengontrol tiga jenis tugas background utama:
1. **Generate Invoices Otomatis**: Secara terjadwal menghasilkan invoice bulanan bagi pelanggan aktif H-5 sebelum tanggal jatuh tempo.
2. **Suspended Otomatis**: Melakukan pengecekan tagihan tertunggak dan mengisolasi koneksi internet pelanggan di MikroTik router.
3. **Verifikasi Pembayaran Otomatis**: Melakukan sinkronisasi pencocokan berkala dengan sistem pembayaran Xendit.

Anda dapat mengubah parameter jadwal berbasis ekspresi Cron (misalnya: `0 0 * * *` untuk berjalan setiap tengah malam) dan langsung menyimpannya tanpa perlu menyentuh kode program.
