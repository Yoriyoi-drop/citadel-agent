# 🛡️ Citadel Agent - Autonomous Secure Workflow Engine

[![Version](https://img.shields.io/badge/version-0.1.0-blue.svg)](https://github.com/citadel-agent/citadel-agent)
[![License](https://img.shields.io/badge/license-Apache2.0-green.svg)](LICENSE)
[![Go](https://img.shields.io/badge/language-Go-lightgrey.svg)](https://golang.org/)
[![TypeScript](https://img.shields.io/badge/language-TypeScript-blue.svg)](https://www.typescriptlang.org/)

> **Citadel Agent** adalah platform otomasi workflow modern dengan kemampuan AI agent, multi-language runtime, dan sandboxing keamanan lanjutan. Platform ini dirancang untuk menggantikan solusi seperti n8n, Windmill, Temporal, dan Prefect dengan integrasi AI agent yang canggih.

## 🚀 Fitur Utama

### 🔐 Keamanan Tingkat Enterprise
- **Sandboxing Node**: Eksekusi node dalam lingkungan terisolasi
- **RBAC (Role-Based Access Control)**: Sistem otorisasi berbasis peran
- **Enkripsi End-to-End**: Perlindungan data sensitif
- **Audit Logging**: Pelacakan aktivitas penuh
- **Policy Isolation**: Pembatasan akses berbasis kebijakan

### 🧠 AI Agent Runtime
- **Memori Agent**: Sistem memori jangka pendek dan panjang
- **Tool Integration**: Integrasi dengan layanan eksternal
- **Multi-Agent Coordination**: Koordinasi antar agent AI
- **Human-in-the-Loop**: Involvement manusia dalam alur AI

### 🌐 Multi-Language Runtime
- **Go Native**: Eksekusi Go dengan kinerja tinggi
- **JavaScript/Node.js**: Eksekusi JavaScript dalam sandbox
- **Python**: Eksekusi Python dalam sandbox aman
- **Java, Ruby, PHP, Rust, C#**: Dukungan bahasa tambahan
- **Container Execution**: Eksekusi dalam container terisolasi

### ⚙️ Workflow Engine
- **Dependency Resolution**: Resolusi dependensi otomatis
- **Parallel Execution**: Eksekusi paralel node
- **Retry Mechanisms**: Strategi retry otomatis
- **Error Handling**: Penanganan error canggih
- **Monitoring**: Pemantauan real-time

### 🧩 Plugin & Node System
- **200+ Pre-built Nodes**: Kategori node dari dasar hingga lanjut
- **Plugin Marketplace**: Marketplace plugin yang aman
- **Custom Node SDK**: Toolkit untuk pembuatan node kustom
- **Version Management**: Manajemen versi plugin
- **Security Scanning**: Pemindaian keamanan otomatis

## 🎯 Use Cases

### 🏢 Enterprise Automation
- Integrasi sistem antar departemen
- Otomasi proses bisnis kompleks
- Pengelolaan workflow yang aman

### 🤖 AI Agent Workflows
- Chatbot cerdas dengan memori
- Asisten otomatis dengan alat eksternal
- Koordinasi tugas-tugas kompleks

### 🌐 Multi-Language Integration
- Eksekusi kode dalam berbagai bahasa
- Integrasi dengan layanan legacy
- Pengolahan data poliglot

## 📊 Arsitektur Sistem

```
┌─────────────────────────────────────────────────────────────────┐
│                    CITADEL-AGENT ARCHITECTURE                   │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐            │
│  │   FRONTEND  │  │   ENGINE    │  │     AI      │            │
│  │    UI       │  │   CORE      │  │   AGENTS    │            │
│  │             │  │             │  │             │            │
│  └─────────────┘  └─────────────┘  └─────────────┘            │
│         │                   │                   │             │
│         ▼                   ▼                   ▼             │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │                  WORKFLOW ENGINE                        │  │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐      │  │
│  │  │   RUNNER    │ │  EXECUTOR   │ │ SCHEDULER   │      │  │
│  │  └─────────────┘ └─────────────┘ └─────────────┘      │  │
│  └─────────────────────────────────────────────────────────┘  │
│                              │                                │
│                              ▼                                │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │                   NODE RUNTIME                          │  │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐      │  │
│  │  │    GO       │ │  JS/TS      │ │ PYTHON      │      │  │
│  │  │  RUNTIME    │ │  SANDBOX    │ │  SANDBOX    │      │  │
│  │  └─────────────┘ └─────────────┘ └─────────────┘      │  │
│  └─────────────────────────────────────────────────────────┘  │
│                              │                                │
│                              ▼                                │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │                    STORAGE LAYER                        │  │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐      │  │
│  │  │   POSTGRES  │ │    REDIS    │ │   FILES     │      │  │
│  │  │             │ │             │ │             │      │  │
│  │  └─────────────┘ └─────────────┘ └─────────────┘      │  │
│  └─────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

## 🛠️ Instalasi

### Persyaratan Sistem
- **OS**: Linux/macOS/Windows
- **Git**: v2.20+
- **Go**: v1.19+
- **Node.js**: v16+
- **Docker**: v20+ (opsional)
- **Docker Compose**: v2+ (opsional)

### Melalui Installer Otomatis
```bash
# Download dan jalankan installer
curl -O https://raw.githubusercontent.com/citadel-agent/citadel-agent/main/install.sh
chmod +x install.sh
./install.sh
```

### Instalasi Manual
```bash
# Clone repositori
git clone https://github.com/citadel-agent/citadel-agent.git
cd citadel-agent

# Pindah ke direktori backend
cd backend

# Install dependensi Go
go mod tidy

# Bangun service API
go build -o ../bin/api cmd/api/main.go

# Pindah ke direktori frontend
cd ../frontend

# Install dependensi Node.js
npm install
```

## 🖥️ Jalankan Citra Terminal

Citadel Agent dilengkapi dengan antarmuka terminal interaktif:

```bash
# Jalankan CLI sederhana
python citadel_cli.py

# Atau versi lanjutan dengan Rich UI
python citadel_advanced_cli.py
```

## 🌐 API Endpoint

Citadel Agent menyediakan API RESTful yang lengkap:

### Autentikasi
```bash
# Login
curl -X POST http://localhost:5001/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@citadel-agent.com", "password":"secure_password"}'
```

### Manajemen Workflow
```bash
# Buat workflow baru
curl -X POST http://localhost:5001/api/v1/workflows \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{...workflow_definition}'
```

### Eksekusi Workflow
```bash
# Jalankan workflow
curl -X POST http://localhost:5001/api/v1/workflows/{id}/run \
  -H "Authorization: Bearer <token>"
```

## 🔐 Konfigurasi Keamanan

### File .env
**PENTING**: File `.env` berisi kredensial dan informasi sensitif, sehingga tidak disertakan dalam repository. Anda perlu membuat file ini sendiri:

1. Buat salinan dari file `.env.example`:
```bash
cp .env.example .env
```

2. Edit file `.env` dan sesuaikan nilai-nilai berikut:
```env
# JWT Secret untuk autentikasi (ganti dengan nilai yang kuat untuk production)
JWT_SECRET=your-super-secret-jwt-key-here-at-least-32-characters-for-production

# Konfigurasi database
DATABASE_URL=postgresql://postgres:password@localhost:5432/citadel_agent

# Konfigurasi Redis
REDIS_URL=redis://localhost:6379

# Environment (development/production)
ENVIRONMENT=development

# Konfigurasi OAuth untuk GitHub dan Google (opsional)
GITHUB_CLIENT_ID=your_github_client_id
GITHUB_CLIENT_SECRET=your_github_client_secret
GITHUB_CALLBACK_URL=http://localhost:5001/api/v1/auth/github/callback

GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
GOOGLE_CALLBACK_URL=http://localhost:5001/api/v1/auth/google/callback

# Batas rate limit API
API_RATE_LIMIT=1000
```

**Catatan**: Jangan pernah mengunggah file `.env` ke repository publik karena berisi informasi sensitif.

### Sandboxing Konfigurasi
Citadel Agent menggunakan sistem sandboxing untuk keamanan:
- **JavaScript**: Dieksekusi dalam vm2 sandbox
- **Python**: Dieksekusi sebagai proses terisolasi
- **Go**: Dieksekusi langsung dengan kontrol sumber daya
- **Container**: Opsi untuk eksekusi dalam container Docker

## 🧪 Testing

### Jalankan Unit Test
```bash
# Backend tests
cd backend
go test ./internal/... -v

# Frontend tests
cd frontend
npm test
```

### Jalankan Integration Test
```bash
# Backend integration tests
cd backend
go test ./test/integration/... -v
```

## 📚 Dokumentasi

### Struktur Dokumentasi
```
docs/
├── api/              # Dokumentasi API
├── guides/           # Panduan penggunaan
├── architecture/     # Dokumentasi arsitektur
└── nodes/            # Dokumentasi node
```

### Panduan Awal
1. **Getting Started**: Mulai dengan Citadel Agent
2. **Workflow Creation**: Membuat workflow pertama
3. **Node Development**: Membangun node kustom
4. **AI Agent Setup**: Konfigurasi agent AI
5. **Security Guide**: Panduan keamanan

## 🚀 Deployment

### Development
```bash
# Jalankan semua service
docker-compose -f docker/compose/docker-compose.dev.yml up -d

# Atau run secara manual
cd backend && go run cmd/api/main.go
```

### Production
```bash
# Deployment dengan Docker Compose
docker-compose -f docker/compose/docker-compose.prod.yml up -d

# Atau deployment dengan Kubernetes
kubectl apply -f k8s/
```

## 🤝 Kontribusi

Kami menyambut kontribusi dari komunitas! Silakan baca [CONTRIBUTING.md](./CONTRIBUTING.md) untuk detail cara berkontribusi.

### Panduan Kontribusi
1. Fork repositori
2. Buat branch fitur (`git checkout -b feature/amazing-feature`)
3. Lakukan perubahan
4. Commit perubahan (`git commit -m 'Add amazing feature'`)
5. Push ke branch (`git push origin feature/amazing-feature`)
6. Buka Pull Request

## 📄 Lisensi

Distributed under the Apache 2.0 License. See [LICENSE](./LICENSE) for more information.

## 🆘 Dukungan

- **Issues**: [GitHub Issues](https://github.com/citadel-agent/citadel-agent/issues)
- **Documentation**: [Citadel Agent Documentation](https://citadel-agent.com/docs)
- **Community**: [Discord Community](https://discord.gg/citadel-agent) (jika tersedia)

---

## 🎮 Tampilan Sistem

Berikut beberapa ilustrasi tampilan sistem Citadel-Agent:

### Login Screen
```
=========================================================
                     CITADEL-AGENT
              Autonomous Secure Workflow Engine
=========================================================

[ AUTHENTICATION REQUIRED ]

 > Username : ________________________________
 > Password : ________________________________

---------------------------------------------------------
  STATUS : Secure channel initialized
  ENGINE : Foundation-Core v0.1.0
  MODE   : Operator Login

  NOTE :
    - Pastikan kredensial benar.
    - Akses ini akan dicatat dalam event-log.
    - Sistem menggunakan sandbox & policy isolation.
---------------------------------------------------------

   Tekan ENTER untuk memulai sesi operasional...
=========================================================
```

**Citadel Agent v0.1.0** - *Platform otomasi workflow generasi berikutnya dengan integrasi AI agent*

---

<p align="center">
  <em>Dibangun untuk organisasi yang mengutamakan keamanan, skalabilitas, dan otomasi cerdas.</em>
</p>