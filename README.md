# Simple Blackjack REST API

Backend REST API sederhana untuk simulasi permainan kartu Blackjack, dibangun menggunakan bahasa pemrograman Go, framework Gin, dan GORM dengan basis data PostgreSQL (Supabase).

Aplikasi ini mencakup manajemen user dengan saldo koin serta siklus hidup permainan Blackjack secara penuh (Create, Read, Update, Delete).

---

## 1. Fitur Utama

- **Otomatisasi Database**: Menggunakan fitur GORM AutoMigrate, tabel `users` dan `games` langsung dibuat otomatis pada PostgreSQL/Supabase saat server pertama kali dijalankan.
- **Autentikasi**: Menggunakan HTTP Basic Auth (username dan password) yang aman menggunakan hashing bcrypt.
- **Manajemen User & Koin**: Registrasi user, pengecekan profil dan saldo, serta top-up koin sebagai mata uang taruhan.
- **Mekanika Blackjack Standar**:
  - Pembagian 2 kartu awal untuk player dan dealer.
  - Kartu kedua dealer disembunyikan (hole card) selama status game masih `IN_PROGRESS`.
  - Hit: Menarik kartu tambahan dari sisa 52 kartu. Jika skor melebihi 21, player langsung dinyatakan Bust (kalah).
  - Stand: Mengakhiri giliran player. Dealer menarik kartu otomatis hingga skor mencapai minimal 17.
  - Penilaian As Dinamis: Nilai As otomatis beralih antara 11 atau 1 untuk mencegah skor bust.
  - Payout: Kemenangan standar membayar 2x taruhan, kemenangan Blackjack alami di awal membayar 2.5x taruhan, dan seri (Push) mengembalikan nilai taruhan penuh.
- **Operasi CRUD Lengkap**:
  - Create: Memulai game baru (`POST /games`).
  - Read: Mengambil riwayat daftar game (`GET /games`) dan detail game (`GET /games/:id`).
  - Update: Melanjutkan giliran kartu (`POST /games/:id/hit` dan `POST /games/:id/stand`).
  - Delete: Menghapus catatan riwayat game tertentu (`DELETE /games/:id`).

---

## 2. Struktur Proyek

```text
simple-blackjack/
|-- config/
|   `-- database.go       # Koneksi PostgreSQL & eksekusi GORM AutoMigrate
|-- controller/
|   |-- game.go           # Logika permainan Blackjack & operasi CRUD game
|   `-- user.go           # Registrasi, profil, dan penambahan koin
|-- models/
|   `-- models.go         # Definisi struct GORM dan fungsi skor kartu
|-- routers/
|   `-- router.go         # Routing Gin & middleware Basic Auth
|-- .env.example          # Template konfigurasi environment
|-- game_logic_test.go    # Unit test evaluasi skor kartu dan kartu As
|-- go.mod                # Daftar dependensi Go
|-- go.sum
|-- main.go               # Titik masuk aplikasi
`-- README.md             # Dokumentasi API
```

---

## 3. Prasyarat dan Konfigurasi

### Prasyarat
- Go versi 1.20 atau yang lebih baru.
- Instance database PostgreSQL lokal atau proyek database di Supabase.

### Langkah Menjalankan Aplikasi

1. Salin template environment:
   ```bash
   cp .env.example .env
   ```

2. Buka berkas `.env` dan sesuaikan URL koneksi database:
   ```env
   DATABASE_URL=postgresql://postgres.[PROJECT_REF]:[PASSWORD]@aws-0-[REGION].pooler.supabase.com:6543/postgres?sslmode=require
   PORT=8080
   ```
   *Catatan: Jika menggunakan PostgreSQL lokal di Laragon/lokal, gunakan format:*
   ```env
   DATABASE_URL=host=localhost user=postgres password=root dbname=simple_blackjack port=5432 sslmode=disable
   ```

3. Jalankan server:
   ```bash
   go run main.go
   ```
   Server akan otomatis membaca `.env`, melakukan migrasi tabel ke database, dan mendengarkan request pada port `8080`.

4. Menjalankan Unit Test:
   ```bash
   go test -v ./...
   ```

### Langkah Deployment ke Vercel

1. Push seluruh repositori ini ke GitHub / GitLab.
2. Buka dashboard Vercel, pilih **Add New Project**, lalu import repositori ini.
3. Pada menu **Environment Variables**, tambahkan:
   - `DATABASE_URL`: URI connection string dari Supabase (pilih mode **Transaction pooler** port 6543 dengan format `postgresql://postgres.[PROJECT_REF]:[PASSWORD]@aws-0-[REGION].pooler.supabase.com:6543/postgres?sslmode=require`).
4. Klik **Deploy**. Vercel akan otomatis mendeteksi server Go `main.go` dan menjalankannya secara native.

---

## 4. Dokumentasi Endpoint API

Semua endpoint dengan keterangan `Perlu Auth` membutuhkan HTTP Basic Auth header dengan kredensial user terdaftar:
`Authorization: Basic <base64(username:password)>` atau via cURL flag `-u username:password`.

### Ringkasan Endpoint

| Method | Endpoint | Auth | Deskripsi |
| :--- | :--- | :--- | :--- |
| POST | /users/register | Tidak | Registrasi akun baru |
| GET | /users/me | Ya | Mendapatkan data profil dan saldo koin |
| POST | /users/coin | Ya | Melakukan penambahan (top-up) koin |
| POST | /games | Ya | Memulai permainan baru dan pasang taruhan |
| GET | /games | Ya | Melihat daftar seluruh riwayat permainan user |
| GET | /games/:id | Ya | Melihat detail status sesi permainan |
| POST | /games/:id/hit | Ya | Menarik satu kartu tambahan |
| POST | /games/:id/stand | Ya | Berhenti menarik kartu dan giliran dealer |
| DELETE | /games/:id | Ya | Menghapus riwayat sesi permainan |

---

### Detail Endpoint & Contoh Request

#### A. Registrasi User
- **URL**: `POST /users/register`
- **Request Body**:
  ```json
  {
    "username": "pemain1",
    "password": "password123",
    "coin": 1000
  }
  ```
- **Response (201 Created)**:
  ```json
  {
    "message": "user registered successfully",
    "user": {
      "id": 1,
      "username": "pemain1",
      "coin": 1000,
      "created_at": "2026-10-30T00:00:00Z"
    }
  }
  ```
- **Contoh cURL**:
  ```bash
  curl -X POST http://localhost:8080/users/register \
    -H "Content-Type: application/json" \
    -d "{\"username\": \"pemain1\", \"password\": \"password123\", \"coin\": 1000}"
  ```

---

#### B. Cek Profil & Saldo Koin
- **URL**: `GET /users/me`
- **Auth**: Basic Auth
- **Response (200 OK)**:
  ```json
  {
    "id": 1,
    "username": "pemain1",
    "coin": 1000,
    "created_at": "2026-10-30T00:00:00Z"
  }
  ```
- **Contoh cURL**:
  ```bash
  curl -X GET http://localhost:8080/users/me \
    -u pemain1:password123
  ```

---

#### C. Top-up Koin
- **URL**: `POST /users/coin`
- **Auth**: Basic Auth
- **Request Body**:
  ```json
  {
    "amount": 500
  }
  ```
- **Response (200 OK)**:
  ```json
  {
    "message": "coin added successfully",
    "added": 500,
    "user_id": 1,
    "new_total": 1500
  }
  ```
- **Contoh cURL**:
  ```bash
  curl -X POST http://localhost:8080/users/coin \
    -u pemain1:password123 \
    -H "Content-Type: application/json" \
    -d "{\"amount\": 500}"
  ```

---

#### D. Memulai Game Baru (Create)
- **URL**: `POST /games`
- **Auth**: Basic Auth
- **Request Body**:
  ```json
  {
    "bet": 100
  }
  ```
- **Response (201 Created)**:
  ```json
  {
    "id": 1,
    "user_id": 1,
    "bet_amount": 100,
    "player_cards": [
      { "suit": "S", "rank": "10", "value": 10 },
      { "suit": "H", "rank": "8", "value": 8 }
    ],
    "dealer_cards": [
      { "suit": "D", "rank": "K", "value": 10 },
      { "suit": "?", "rank": "?", "value": 0 }
    ],
    "player_score": 18,
    "dealer_score": 10,
    "status": "IN_PROGRESS",
    "created_at": "2026-10-30T00:05:00Z",
    "updated_at": "2026-10-30T00:05:00Z"
  }
  ```
- **Contoh cURL**:
  ```bash
  curl -X POST http://localhost:8080/games \
    -u pemain1:password123 \
    -H "Content-Type: application/json" \
    -d "{\"bet\": 100}"
  ```

---

#### E. Hit (Ambil Kartu Tambahan) (Update)
- **URL**: `POST /games/:id/hit`
- **Auth**: Basic Auth
- **Response (200 OK)**:
  ```json
  {
    "id": 1,
    "user_id": 1,
    "bet_amount": 100,
    "player_cards": [
      { "suit": "S", "rank": "10", "value": 10 },
      { "suit": "H", "rank": "8", "value": 8 },
      { "suit": "C", "rank": "2", "value": 2 }
    ],
    "dealer_cards": [
      { "suit": "D", "rank": "K", "value": 10 },
      { "suit": "?", "rank": "?", "value": 0 }
    ],
    "player_score": 20,
    "dealer_score": 10,
    "status": "IN_PROGRESS",
    "created_at": "2026-10-30T00:05:00Z",
    "updated_at": "2026-10-30T00:06:00Z"
  }
  ```
- **Contoh cURL**:
  ```bash
  curl -X POST http://localhost:8080/games/1/hit \
    -u pemain1:password123
  ```

---

#### F. Stand (Selesaikan Giliran) (Update)
- **URL**: `POST /games/:id/stand`
- **Auth**: Basic Auth
- **Response (200 OK)**:
  ```json
  {
    "id": 1,
    "user_id": 1,
    "bet_amount": 100,
    "player_cards": [
      { "suit": "S", "rank": "10", "value": 10 },
      { "suit": "H", "rank": "8", "value": 8 },
      { "suit": "C", "rank": "2", "value": 2 }
    ],
    "dealer_cards": [
      { "suit": "D", "rank": "K", "value": 10 },
      { "suit": "C", "rank": "7", "value": 7 }
    ],
    "player_score": 20,
    "dealer_score": 17,
    "status": "PLAYER_WON",
    "created_at": "2026-10-30T00:05:00Z",
    "updated_at": "2026-10-30T00:07:00Z"
  }
  ```
- **Contoh cURL**:
  ```bash
  curl -X POST http://localhost:8080/games/1/stand \
    -u pemain1:password123
  ```

---

#### G. Lihat Daftar Seluruh Permainan (Read)
- **URL**: `GET /games`
- **Auth**: Basic Auth
- **Response (200 OK)**:
  ```json
  {
    "total": 1,
    "games": [
      {
        "id": 1,
        "user_id": 1,
        "bet_amount": 100,
        "player_cards": [...],
        "dealer_cards": [...],
        "player_score": 20,
        "dealer_score": 17,
        "status": "PLAYER_WON",
        "created_at": "2026-10-30T00:05:00Z",
        "updated_at": "2026-10-30T00:07:00Z"
      }
    ]
  }
  ```
- **Contoh cURL**:
  ```bash
  curl -X GET http://localhost:8080/games \
    -u pemain1:password123
  ```

---

#### H. Lihat Detail Permainan (Read)
- **URL**: `GET /games/:id`
- **Auth**: Basic Auth
- **Response (200 OK)**:
  ```json
  {
    "id": 1,
    "user_id": 1,
    "bet_amount": 100,
    "player_cards": [...],
    "dealer_cards": [...],
    "player_score": 20,
    "dealer_score": 17,
    "status": "PLAYER_WON",
    "created_at": "2026-10-30T00:05:00Z",
    "updated_at": "2026-10-30T00:07:00Z"
  }
  ```
- **Contoh cURL**:
  ```bash
  curl -X GET http://localhost:8080/games/1 \
    -u pemain1:password123
  ```

---

#### I. Hapus Riwayat Permainan (Delete)
- **URL**: `DELETE /games/:id`
- **Auth**: Basic Auth
- **Response (200 OK)**:
  ```json
  {
    "message": "game history deleted successfully",
    "game_id": 1
  }
  ```
- **Contoh cURL**:
  ```bash
  curl -X DELETE http://localhost:8080/games/1 \
    -u pemain1:password123
  ```
