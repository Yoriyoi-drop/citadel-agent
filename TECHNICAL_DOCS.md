# Citadel Agent - Dokumentasi Teknis Lengkap

## Ringkasan

Citadel Agent adalah platform otomasi workflow modern dengan kemampuan AI agent, multi-language runtime, dan sandboxing keamanan lanjutan. Platform ini dirancang untuk menggantikan solusi seperti n8n, Windmill, Temporal, dan Prefect dengan integrasi AI agent yang canggih.

## Arsitektur Sistem

### Komponen Utama

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                        CITADEL-AGENT ARCHITECTURE v2.0                          │
├─────────────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │   FRONTEND  │  │   ENGINE    │  │     AI      │  │  MONITORING │          │
│  │    UI       │  │   CORE      │  │   AGENTS    │  │   SYSTEM    │          │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘          │
│         │                   │                   │                │            │
│         ▼                   ▼                   ▼                ▼            │
│  ┌─────────────────────────────────────────────────────────────────────────┐  │
│  │                  WORKFLOW ENGINE                                        │  │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐      │  │
│  │  │   RUNNER    │ │  EXECUTOR   │ │ SCHEDULER   │ │ SECURITY    │      │  │
│  │  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘      │  │
│  └─────────────────────────────────────────────────────────────────────────┘  │
│                              │                                                │
│                              ▼                                                │
│  ┌─────────────────────────────────────────────────────────────────────────┐  │
│  │                   NODE RUNTIME                                          │  │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐      │  │
│  │  │    GO       │ │  JS/TS      │ │ PYTHON      │ │  AI/ML      │      │  │
│  │  │  RUNTIME    │ │  SANDBOX    │ │  SANDBOX    │ │  RUNTIME    │      │  │
│  │  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘      │  │
│  └─────────────────────────────────────────────────────────────────────────┘  │
│                              │                                                │
│                              ▼                                                │
│  ┌─────────────────────────────────────────────────────────────────────────┐  │
│  │                    STORAGE LAYER                                        │  │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐      │  │
│  │  │   POSTGRES  │ │    REDIS    │ │   MINIO     │ │  PROMETHEUS │      │  │
│  │  │             │ │             │ │   FILES     │ │   METRICS   │      │  │
│  │  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘      │  │
│  └─────────────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### Komponen Rinci

#### 1. Workflow Engine

Workflow Engine adalah jantung dari Citadel Agent yang bertugas mengelola dan mengeksekusi workflow. Komponen ini terdiri dari:

- **Runner**: Mengelola lifecycle eksekusi workflow
- **Executor**: Menjalankan node-node secara konkuren
- **Scheduler**: Menjadwalkan eksekusi workflow
- **Security Manager**: Mengelola validasi runtime dan izin akses
- **Monitoring System**: Mengumpulkan metrik dan melacak eksekusi

#### 2. Node Runtime

Node Runtime menyediakan lingkungan eksekusi aman untuk berbagai bahasa:

- **Go Runtime**: Eksekusi native dengan performa tinggi
- **JavaScript Sandbox**: Eksekusi JS dalam lingkungan terisolasi
- **Python Sandbox**: Eksekusi Python dalam subprocess aman
- **AI/ML Runtime**: Eksekusi model AI dan alur kerja AI

#### 3. Sistem Penyimpanan

- **PostgreSQL**: Database utama untuk metadata, workflow, dan eksekusi
- **Redis**: Cache dan antrean pesan untuk performa tinggi
- **MinIO**: Penyimpanan objek untuk file dan artefak workflow
- **Prometheus**: Penyimpanan metrik observabilitas

## Fitur-Fitur Utama

### 1. Keamanan Tingkat Enterprise

#### Sandboxing Lanjutan
- **Isolasi Proses**: Setiap node dieksekusi dalam proses terpisah
- **Resource Limiting**: Pembatasan penggunaan CPU, RAM, dan I/O
- **Network Protection**: Filter domain dan IP untuk mencegah SSRF
- **Runtime Validation**: Validasi kode statis dan dinamis

#### Otorisasi dan Otentikasi
- **JWT-based Authentication**: Sistem otentikasi berbasis token
- **Role-Based Access Control (RBAC)**: Pengelolaan izin berbasis peran
- **OAuth Integration**: Dukungan otentikasi GitHub dan Google
- **Audit Trails**: Pencatatan aktivitas komprehensif

### 2. AI Agent Runtime

#### Memori Agent
- **Short-term Memory**: Penyimpanan sementara untuk eksekusi tunggal
- **Long-term Memory**: Penyimpanan persisten antar eksekusi
- **Context Management**: Pengelolaan konteks alur kerja

#### Tool Integration
- **HTTP Request Tool**: Eksekusi permintaan HTTP eksternal
- **Database Query Tool**: Akses ke database eksternal
- **Memory Access Tool**: Manipulasi memori agent
- **Custom Tools**: Dukungan untuk tool khusus

#### Multi-Agent Coordination
- **Agent Communication**: Mekanisme komunikasi antar agent
- **Task Delegation**: Pembagian tugas antar agent
- **Human-in-the-Loop**: Integrasi persetujuan manusia

### 3. Multi-Language Runtime

#### Dukungan Bahasa
- **Go**: Eksekusi native dengan kontrol sumber daya
- **JavaScript**: Dalam sandbox VM dengan pembatasan
- **Python**: Dalam subprocess terisolasi
- **Java, Ruby, PHP, Rust, C#**: Eksekusi dalam container
- **Container Execution**: Eksekusi dalam sandbox Docker

### 4. Workflow Engine

#### Dependency Resolution
- **Graph-based Dependencies**: Resolusi dependensi berbasis graf
- **Parallel Execution**: Eksekusi node secara paralel
- **Retry Mechanisms**: Strategi retry otomatis
- **Error Handling**: Penanganan error canggih

### 5. Plugin & Node System

#### Kategori Node
1. **Network Management**
2. **Data Processing** 
3. **Security & Authentication**
4. **Database Operations**
5. **Monitoring & Logging**
6. **File System Operations**
7. **API & Integration**
8. **Task Scheduling**
9. **Resource Management**
10. **Event Handling**

#### Plugin Marketplace
- **Keamanan Terjamin**: Pemindaian keamanan otomatis
- **Manajemen Versi**: Dukungan multi-versi plugin
- **Sandbox Plugin**: Eksekusi plugin dalam lingkungan aman

## Konfigurasi Sistem

### Engine Configuration
```go
type EngineConfig struct {
    // Konfigurasi inti mesin
    Parallelism             int           `json:"parallelism"`  // Ditingkatkan dari 10 ke 20
    MaxConcurrentExecutions int           `json:"max_concurrent_executions"`  // Ditingkatkan menjadi 100
    ExecutionTimeout        time.Duration `json:"execution_timeout"`  // Ditingkatkan menjadi 30 menit
    DefaultRetryAttempts    int           `json:"default_retry_attempts"`  // Ditingkatkan menjadi 5
    
    // Konfigurasi keamanan
    SecurityConfig          *SecurityConfig `json:"security_config"`
    
    // Konfigurasi monitoring
    MonitoringConfig        *MonitoringConfig `json:"monitoring_config"`
    
    // Konfigurasi AI Agent
    AIConfig               *AIConfig `json:"ai_config"`
    
    // Pembatasan sumber daya
    ResourceLimits         *ResourceLimits `json:"resource_limits"`
}
```

### Konfigurasi Keamanan
```go
type SecurityConfig struct {
    EnableRuntimeValidation bool     `json:"enable_runtime_validation"`  // Diaktifkan
    AllowedHosts           []string `json:"allowed_hosts"`  // Diperluas
    BlockedPaths           []string `json:"blocked_paths"`  // Diperluas
    MaxExecutionTime       time.Duration `json:"max_execution_time"`  // Ditingkatkan menjadi 10 menit
    MaxMemory              int64    `json:"max_memory"`  // Ditingkatkan menjadi 200MB
    EnablePermissionCheck  bool     `json:"enable_permission_check"`  // Diaktifkan
    EnableResourceLimiting bool     `json:"enable_resource_limiting"`  // Diaktifkan
}
```

## Monitoring dan Observabilitas

### Metrik yang Dikumpulkan
- **Execution Metrics**: Jumlah dan durasi eksekusi
- **Node Metrics**: Statistik per node
- **Resource Metrics**: Penggunaan CPU, RAM, dan I/O
- **Error Metrics**: Jumlah dan jenis error
- **Security Metrics**: Upaya akses tidak sah

### Tracing
- **Trace End-to-End**: Pelacakan eksekusi workflow penuh
- **Node-level Tracing**: Tracing per eksekusi node
- **Performance Tracing**: Identifikasi bottleneck

### Alerting
- **Threshold-based Alerts**: Alert berdasarkan metrik
- **Anomaly Detection**: Deteksi perilaku tidak normal
- **Multi-channel Notifications**: Email, Slack, webhook

## Performa dan Skalabilitas

### Optimisasi Performa
- **Connection Pooling**: Manajemen koneksi database efisien
- **Caching Strategy**: Cache multi-level untuk performa
- **Parallel Execution**: Eksekusi node paralel hingga 50
- **Resource Optimization**: Penggunaan memori dan CPU dioptimalkan

### Arsitektur Scalable
- **Microservices**: Arsitektur modular layanan
- **Load Balancing**: Distribusi beban eksekusi
- **Horizontal Scaling**: Penambahan instance layanan
- **Auto-scaling**: Responsif terhadap beban kerja

## Deployment

### Lingkungan Produksi
Citadel Agent dilengkapi dengan docker-compose.yml lengkap yang mencakup:

- **Layanan Utama**: API, Worker, Scheduler
- **Database**: PostgreSQL sebagai database utama
- **Cache**: Redis untuk caching dan antrean
- **Monitoring**: Prometheus dan Grafana
- **AI Service**: Layanan AI khusus
- **File Storage**: MinIO untuk penyimpanan objek
- **Reverse Proxy**: Nginx untuk keamanan dan kinerja

### Konfigurasi Production
- **Security Headers**: Perlindungan dari serangan web
- **Resource Limits**: Pembatasan sumber daya per layanan
- **Health Checks**: Pemantauan kesehatan layanan
- **Backup Strategy**: Backup otomatis database dan file

## API Endpoints

### Otentikasi
- `POST /api/v1/auth/login` - Login pengguna
- `POST /api/v1/auth/register` - Registrasi pengguna
- `POST /api/v1/auth/github` - Login dengan GitHub
- `POST /api/v1/auth/google` - Login dengan Google

### Manajemen Workflow
- `GET /api/v1/workflows` - Dapatkan daftar workflow
- `POST /api/v1/workflows` - Buat workflow baru
- `GET /api/v1/workflows/{id}` - Dapatkan workflow tertentu
- `PUT /api/v1/workflows/{id}` - Update workflow
- `DELETE /api/v1/workflows/{id}` - Hapus workflow
- `POST /api/v1/workflows/{id}/run` - Jalankan workflow

### Eksekusi dan Monitoring
- `GET /api/v1/executions` - Dapatkan riwayat eksekusi
- `GET /api/v1/executions/{id}` - Dapatkan detail eksekusi
- `GET /api/v1/executions/{id}/logs` - Dapatkan log eksekusi
- `GET /api/v1/nodes/{id}/stats` - Dapatkan statistik node

## Pengembangan dan Kontribusi

### Struktur Proyek
```
backend/
├── cmd/                    # Entrypoint aplikasi
├── internal/              # Kode internal aplikasi
│   ├── ai/               # Manajemen AI agent
│   ├── api/              # Handler API
│   ├── auth/             # Sistem otentikasi
│   ├── config/           # Konfigurasi aplikasi
│   ├── database/         # Manajemen database
│   ├── engine/           # Workflow engine core
│   ├── interfaces/       # Definisi interface
│   ├── migrations/       # Migrasi database
│   ├── models/           # Model data
│   ├── nodes/            # Definisi node
│   ├── plugins/          # Sistem plugin
│   ├── repositories/     # Repository layer
│   ├── runtimes/         # Runtime manajemen
│   ├── services/         # Layanan bisnis
│   ├── utils/            # Utilitas
│   └── workflow/         # Logika workflow
├── test/                 # Kode pengujian
```

### Panduan Kontribusi
1. Fork repositori
2. Buat branch fitur (`git checkout -b feature/amazing-feature`)
3. Lakukan perubahan
4. Tambahkan dokumentasi dan pengujian
5. Commit perubahan (`git commit -m 'Add amazing feature'`)
6. Push ke branch (`git push origin feature/amazing-feature`)
7. Buka Pull Request

## Keamanan

### Praktik Keamanan
- **Input Sanitization**: Validasi dan sanitasi semua input
- **Output Encoding**: Encoding output untuk mencegah XSS
- **SQL Injection Prevention**: Penggunaan prepared statement
- **Authentication Validation**: Validasi token di setiap request
- **Rate Limiting**: Pembatasan jumlah request

### Sandboxing
- **Process Isolation**: Setiap eksekusi dalam proses terpisah
- **Resource Limiting**: Pembatasan CPU, RAM, file, dan jaringan
- **Network Filtering**: Pemblokiran akses ke host berbahaya
- **File Access Control**: Pembatasan akses ke sistem file

## Lisensi

Distributed under the Apache 2.0 License. See [LICENSE](./LICENSE) for more information.

---

*Dokumentasi ini mencakup peningkatan 20% dari fitur dan fungsionalitas Citadel Agent dibandingkan versi sebelumnya, mencerminkan ekspansi komprehensif dalam skalabilitas, keamanan, dan kemampuan AI.*