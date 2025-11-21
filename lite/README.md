# Citadel Agent Lite

Citadel Agent versi ringan dan stabil dengan fokus pada autentikasi dan workflow sederhana.

## Fitur

- Otentikasi lokal (email/password)
- OAuth GitHub dan Google
- API endpoint sederhana
- Konfigurasi mudah
- Ringan dan cepat

## Arsitektur Sederhana

```
Citadel Agent Lite
├── main.go              # Aplikasi utama
├── go.mod              # Dependencies ringan
├── Dockerfile          # Untuk deployment
├── .env                # Konfigurasi
└── README.md
```

## Cara Menjalankan

### Lokal
```bash
# Install dependencies
go mod tidy

# Set environment variables
export JWT_SECRET=your_secret
export DATABASE_URL=postgresql://user:pass@localhost:5432/db

# Jalankan aplikasi
go run main.go
```

### Docker
```bash
# Build image
docker build -t citadel-agent-lite .

# Run container
docker run -p 5001:5001 -e JWT_SECRET=your_secret citadel-agent-lite
```

## Endpoint API

- `GET /` - Halaman utama
- `GET /health` - Status kesehatan
- `POST /auth/login` - Login lokal
- `GET /auth/github` - Redirect ke GitHub OAuth
- `GET /auth/github/callback` - Callback dari GitHub
- `GET /auth/google` - Redirect ke Google OAuth
- `GET /auth/google/callback` - Callback dari Google
- `GET /auth/me` - Get informasi user terotentikasi

## Konfigurasi

| Variabel | Default | Keterangan |
|----------|---------|------------|
| `PORT` | `5001` | Port aplikasi |
| `JWT_SECRET` | `default...` | Secret untuk JWT |
| `DATABASE_URL` | `postgresql://...` | Koneksi database |
| `GITHUB_CLIENT_ID` | - | Client ID GitHub OAuth |
| `GITHUB_CLIENT_SECRET` | - | Client Secret GitHub OAuth |
| `GITHUB_REDIRECT_URI` | `http://...` | Callback URL GitHub |
| `GOOGLE_CLIENT_ID` | - | Client ID Google OAuth |
| `GOOGLE_CLIENT_SECRET` | - | Client Secret Google OAuth |
| `GOOGLE_REDIRECT_URI` | `http://...` | Callback URL Google |

## Dependensi Ringan

- `github.com/gofiber/fiber/v2` - Web framework cepat
- `github.com/golang-jwt/jwt/v5` - JWT untuk autentikasi
- `golang.org/x/oauth2` - OAuth 2.0 support
- `github.com/jackc/pgx/v5` - PostgreSQL driver

## Keunggulan Versi Lite

- **Sederhana** - Hanya 1 file utama (main.go)
- **Ringan** - Hanya 4 dependencies utama
- **Cepat** - Fiber framework yang cepat
- **Stabil** - Kode minimal mengurangi potensi error
- **Mudah dimaintain** - Struktur file yang jelas