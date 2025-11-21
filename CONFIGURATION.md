# Konfigurasi Citadel Agent

File `.env` berisi konfigurasi lingkungan untuk Citadel Agent. Anda perlu membuat file ini sendiri berdasarkan contoh berikut.

## Cara Membuat File Konfigurasi

1. Buat salinan file `.env` berdasarkan struktur di bawah:
```bash
# Buat file baru bernama .env
touch .env
```

2. Isi file `.env` dengan variabel lingkungan berikut:

### Konfigurasi Utama
```env
# Server configuration
SERVER_PORT=5001
API_VERSION=v1
ENVIRONMENT=development

# Security configuration (CHANGE THESE FOR PRODUCTION!)
JWT_SECRET=your-super-secret-jwt-key-here-change-in-production
JWT_EXPIRY=86400

# Database configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=citadel_agent

# Redis configuration
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# Logging configuration
LOG_LEVEL=info
LOG_OUTPUT=stdout

# CORS configuration
ALLOWED_ORIGINS=*

# OAuth Configuration for GitHub and Google (optional)
GITHUB_CLIENT_ID=your_github_client_id
GITHUB_CLIENT_SECRET=your_github_client_secret
GITHUB_CALLBACK_URL=http://localhost:5001/api/v1/auth/github/callback

GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
GOOGLE_CALLBACK_URL=http://localhost:5001/api/v1/auth/google/callback
```

### Penjelasan Variabel

- `SERVER_PORT`: Port tempat aplikasi berjalan (default: 5001)
- `JWT_SECRET`: Kunci rahasia untuk enkripsi token (gunakan nilai acak yang kuat untuk production)
- `DB_*`: Kredensial database PostgreSQL
- `REDIS_*`: Koneksi ke server Redis
- `GITHUB_*` dan `GOOGLE_*`: Kredensial OAuth untuk otentikasi GitHub/Google

## Lingkungan Development vs Production

Untuk lingkungan yang berbeda, Anda bisa membuat file konfigurasi yang berbeda:
- `.env` - untuk lingkungan default
- `.env.local` - untuk konfigurasi lokal development
- `.env.production` - untuk lingkungan production

## Keamanan

⚠️ **PERINGATAN KEAMANAN**:
- Jangan pernah menyimpan file `.env` di repository publik
- Gunakan kunci yang kuat dan acak untuk JWT_SECRET
- Jangan gunakan nilai default di lingkungan production
- Perlakukan file `.env` sebagai file sensitif seperti kunci privat