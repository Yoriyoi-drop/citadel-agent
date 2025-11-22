# Citadel Agent

<p align="center">
  <img src="logo.png" alt="Citadel Agent Logo" width="120" height="120">
</p>

<h3 align="center">AI-Powered Workflow Automation Engine</h3>

<p align="center">
  <strong>Advanced orchestration platform with modular node system, visual workflow builder, and AI integration</strong>
</p>

<p align="center">
  <a href="#features">Features</a> •
  <a href="#architecture">Architecture</a> •
  <a href="#installation">Installation</a> •
  <a href="#usage">Usage</a> •
  <a href="#contributing">Contributing</a>
</p>

---

Citadel Agent adalah platform otomasi workflow sumber terbuka yang kuat yang menyatukan kecerdasan buatan, database, kriptografi, dan sistem terdistribusi dalam mesin orkestrasi visual. Dengan sistem node modular dan antarmuka visual yang intuitif, Citadel Agent memungkinkan pengembang dan non-pengembang untuk membuat, menyebarkan, dan mengelola workflow kompleks dengan mudah.

## ✨ Features

- **Visual Workflow Builder** - Drag-and-drop interface untuk membangun workflow kompleks
- **Modular Node System** - Lebih dari 40 jenis node yang dapat dengan mudah disambungkan
- **AI Integration** - Dukungan untuk model AI lanjutan termasuk LLM, Computer Vision, dan NLP
- **Security First** - Built-in security features termasuk encryption, access control, dan OAuth2
- **Database Connectors** - Integrasi native dengan PostgreSQL, MySQL, MongoDB, dan lainnya
- **Real-time Execution** - Eksekusi workflow real-time dengan log dan monitoring
- **Scalable Architecture** - Arsitektur yang dapat diskalakan untuk enterprise
- **Plugin System** - Ekstensibilitas melalui sistem plugin yang fleksibel

## 🏗️ Architecture

Citadel Agent dibangun dengan arsitektur modular yang memisahkan concerns dan memungkinkan ekstensibilitas:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   Workflow      │    │   Node Engine   │
│   (React)       │    │   Engine        │    │   Core          │
│   ┌───────────┐ │    │   ┌───────────┐ │    │   ┌───────────┐ │
│   │ ReactFlow │ │◄──►│   │ Orchestrator│ │◄──►│   │ Node      │ │
│   │ Visual    │ │    │   │ Core        │ │    │   │ Registry│ │
│   │ Builder   │ │    │   │             │ │    │   │         │ │
│   └───────────┘ │    │   └───────────┘ │    │   └───────────┘ │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Node Categories

Citadel Agent menyediakan berbagai kategori node untuk setiap kebutuhan workflow:

#### 🤖 Elite AI Nodes
- **Vision AI Processor** - Image recognition dan computer vision
- **Speech-to-Text/Text-to-Speech** - Audio processing dan synthesis
- **Advanced ML Models** - Training dan inference model ML
- **Natural Language Processor** - Pemrosesan bahasa lanjutan
- **Anomaly Detection AI** - Deteksi anomali dan outlier

#### 🔐 Security Nodes
- **Firewall Manager** - Pengelolaan aturan firewall dan akses
- **Encryption/Decryption** - Enkripsi data dengan berbagai algoritma
- **Access Control** - RBAC dan integrasi LDAP/Active Directory
- **API Key Manager** - Manajemen dan validasi API keys
- **JWT Handler** - Penanganan token JWT
- **OAuth2 Provider** - Server OAuth2 untuk authentication

#### ⚙️ Core & Utility Nodes
- **HTTP Request** - HTTP client untuk integrasi API
- **Data Transformer** - Transformasi dan mapping data
- **Logger** - Logging dengan berbagai output
- **Config Manager** - Manajemen konfigurasi
- **Validator** - Validasi data dengan aturan kompleks
- **UUID Generator** - Pembuatan UUID unik

#### 📊 Database & ORM Nodes
- **GORM/Bun/Ent** - ORM untuk berbagai database
- **SQLC Integration** - SQL compilation dan eksekusi
- **Migrate** - Manajemen skema database

#### 🔄 Workflow & Scheduling
- **Cron Scheduler** - Penjadwalan berbasis Cron
- **Task Queue** - Queue management untuk task async
- **Job Scheduler** - Penjadwalan task cerdas
- **Worker Pool** - Manajemen pool worker
- **Circuit Breaker** - Resiliensi dan fault tolerance

#### 🔌 Integration Nodes
- **REST API Client** - Klien untuk API eksternal
- **AWS S3 Manager** - Integrasi penyimpanan cloud
- **Slack Messenger** - Notifikasi dan komunikasi

## 🚀 Installation

### Prerequisites
- Go 1.21+
- PostgreSQL 12+ (opsional)
- Redis (opsional untuk caching)
- Node.js 18+ (untuk frontend)

### Quick Start

1. Clone repository:
```bash
git clone https://github.com/citadel-agent/citadel-agent.git
cd citadel-agent
```

2. Install dependencies:
```bash
# Backend dependencies
go mod tidy

# Frontend dependencies (if applicable)
cd frontend && npm install
```

3. Configure environment:
```bash
cp .env.example .env
# Edit .env with your configuration
```

4. Run the application:
```bash
go run cmd/server/main.go
```

## 📖 Usage

### Creating Your First Workflow

1. Akses dashboard Citadel Agent
2. Klik "Create New Workflow"
3. Gunakan drag-and-drop untuk menambahkan node
4. Sambungkan node dengan mengklik port mereka
5. Konfigurasi parameter setiap node
6. Simpan dan eksekusi workflow

### Example Workflow: Data Processing Pipeline

```yaml
workflow:
  name: "Data Processing Pipeline"
  nodes:
    - id: http_source
      type: http_request
      config:
        url: "https://api.example.com/data"
        method: "GET"
    
    - id: validator
      type: validator
      config:
        struct_tags:
          email: "required,email"
          age: "required,min=18"
      inputs:
        from: "http_source.output"
    
    - id: processor
      type: utility
      config:
        operation: "string_operation"
        string_operation: "to_upper"
      inputs:
        from: "validator.result"
    
    - id: storage
      type: database_query
      config:
        query: "INSERT INTO processed_data (data) VALUES (?)"
        params: ["processor.result"]
```

## 🛠️ Development

### Adding New Node Types

1. Buat file node baru di direktori kategori yang sesuai:
```go
// backend/internal/nodes/category/new_node.go
package category

import (
    "context"
    "citadel-agent/backend/internal/workflow/core/engine"
)

// NewNodeConfig represents configuration for new node
type NewNodeConfig struct {
    // Define your config fields
}

// NewNode represents the node
type NewNode struct {
    config *NewNodeConfig
}

// Implement NodeInstance interface
func (n *NewNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
    // Your logic here
    return map[string]interface{}{
        "success": true,
        "result":  "result",
    }, nil
}

// Register in node registry
func RegisterNewNode(registry *engine.NodeRegistry) {
    registry.RegisterNodeType("new_node", func(config map[string]interface{}) (engine.NodeInstance, error) {
        return NewNodeFromConfig(config)
    })
}
```

2. Daftarkan node ke registry:
```go
// backend/internal/nodes/registry.go
// Tambahkan tipe node baru dan registrasinya
```

3. Buat UI component untuk node (opsional)

## 🔒 Security

Citadel Agent dirancang dengan keamanan sebagai prioritas:

- **Sandboxing** - Eksekusi kode dalam lingkungan terisolasi
- **Node Isolation** - Pembatasan resource antar node
- **Authentication** - Sistem OAuth2 dan JWT bawaan
- **Authorization** - RBAC dan sistem access control
- **Encryption** - Enkripsi data di transit dan di rest
- **Audit Logging** - Rekaman komprehensif untuk compliance

## 🤝 Contributing

Kami menyambut kontribusi dari komunitas! Untuk berkontribusi:

1. Fork repository
2. Buat branch fitur (`git checkout -b feature/amazing-feature`)
3. Commit perubahan (`git commit -m 'Add amazing feature'`)
4. Push ke branch (`git push origin feature/amazing-feature`)
5. Buka Pull Request

Silakan lihat `CONTRIBUTING.md` untuk panduan kontribusi lengkap.

## 📄 License

Distributed under the Apache License 2.0. See `LICENSE` for more information.

## ⭐ Support

Jika Anda menyukai proyek ini, jangan lupa untuk memberikan ⭐ star! Setiap dukungan sangat berarti.

---

<p align="center">
Made with ❤️ for the open-source community
</p>

<p align="center">
<a href="https://github.com/citadel-agent/citadel-agent/stargazers"><img src="https://img.shields.io/github/stars/citadel-agent/citadel-agent?style=social"></a>
</p>