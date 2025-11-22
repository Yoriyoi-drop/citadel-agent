# Prompt untuk Icon Set Node Citadel-Agent

Gunakan prompt ini untuk DALL-E, Midjourney, atau AI image generation lainnya untuk membuat icon set konsisten untuk semua node Citadel-Agent.

## 🎨 Gaya Umum Icon

```
Create a consistent icon set for Citadel-Agent workflow nodes.
Style: clean outline icons, similar to Lucide Icons or Tabler Icons.
Characteristics: 
- Thin, consistent stroke width (2px)
- Monochromatic color scheme (tech blue/cyan)
- 24x24px or 48x48px square canvas
- Minimal, geometric design
- Recognizable at small sizes
- Flat design with no shadows
- Transparent background
```

## 📦 Icon untuk Setiap Kategori Node

### 🤖 Elite AI Nodes
```
Generate outline-style icons for Elite AI nodes:

- Vision AI Processor: camera with neural network mesh overlay
- Speech-to-Text: microphone with sound waves and text conversion
- Text-to-Speech: speaker with sound waves and text input
- Contextual Reasoning: brain with connected nodes
- Anomaly Detection AI: warning triangle with neural network
- Prediction Model: chart graph with AI symbol
- Sentiment Analysis: face with emotional indicators and data points
- AI-Agent-Orchestrator: multiple connected AI icons
- ML-Model-Training: gear with brain symbol
- Advanced-ML-Inference: computer chip with neural network
- Multi-Modal-AI-Processor: multiple input types (text, image, audio)
- Advanced-Natural-Language-Processor: speech bubble with processing symbol
- Real-time-ML-Training: lightning bolt with gear
- Advanced-Recommendation-Engine: thumbs up with connected dots
- Advanced-AI-Agent-Manager: multiple AI agents with management symbol
- Advanced-Decision-Engine: decision tree graph
- Advanced-Predictive-Analytics: data points with prediction arrow
- Advanced-Content-Intelligence: content blocks with AI symbol
- Advanced-Data-Intelligence: database with AI overlay
```

### 🔐 Security Nodes
```
Generate outline-style icons for Security nodes:

- Firewall Manager: shield with firewall bars
- Encryption: lock with encryption key
- Access Control: key with user control symbol
- API Key Manager: key with API symbol
- JWT Handler: shield with token symbol
- OAuth2 Provider: user authentication with security shield
- Security Operations: multiple security symbols combined
```

### ⚙️ Core & HTTP Nodes
```
Generate outline-style icons for Core nodes:

- HTTP Request: network connection with arrows
- Validator: checkmark with validation symbol
- Logger: document with logging indicator
- Config Manager: settings gear with configuration
- UUID Generator: number sequence with unique identifier
```

### 📊 Database & ORM Nodes
```
Generate outline-style icons for Database nodes:

- GORM Database: database with GORM symbol
- Bun Database: database with Bun symbol
- Ent Database: database with Ent symbol
- SQLC Database: database with SQL compiler symbol
- Migrate Database: database with migration arrows
```

### 🔄 Workflow & Scheduling Nodes
```
Generate outline-style icons for Workflow nodes:

- Cron Scheduler: clock with cron symbol
- Task Queue: queue line with task cards
- Job Scheduler: calendar with job symbol
- Worker Pool: multiple user icons with work symbol
- Circuit Breaker: electrical circuit with break indicator
```

### 🔧 Debug & Logging Nodes
```
Generate outline-style icons for Debug nodes:

- Debug Node: bug with debug symbol
- Logging: log file with timestamp
```

### 🔧 Utility Nodes
```
Generate outline-style icons for Utility nodes:

- Utility Node: toolbox with various tools
```

### 🔲 Basic Nodes
```
Generate outline-style icons for Basic nodes:

- Basic Node: simple geometric shape
- Constant: equal sign or constant symbol
- Passthrough: arrows continuing straight
- Delay: clock with delay symbol
- Counter: numbers or count symbol
- Condition: if/then symbol
- Loop: circular arrow
- Switch: toggle or switch symbol
- Math: mathematical operations
```

### 🔌 Plugin Nodes
```
Generate outline-style icons for Plugin nodes:

- Plugin: puzzle piece with connection
- HTTP Plugin: plugin with network
- Database Plugin: plugin with database
- Message Queue Plugin: plugin with queue
- Storage Plugin: plugin with storage
- External Service Plugin: plugin with external connection
```

## 🎯 Panduan Tambahan

### Warna:
- Gunakan warna konsisten: `#2563EB` (biru teknologi) atau `#06B6D4` (cyan)
- Variasi untuk kategori: Security (merah), AI (ungu), Data (hijau), Utility (kuning)

### Konsistensi:
- Pastikan semua icon memiliki stroke width yang konsisten
- Gunakan sudut tajam daripada lekukan untuk tampilan teknologi
- Buat ruang negatif yang seimbang di dalam ikon
- Pastikan ikon dikenali dari jarak jauh

### Penggunaan:
- Ikon akan digunakan di workflow builder untuk mengidentifikasi tipe node
- Ikon harus mudah dibedakan antara satu sama lain
- Desain agar mudah diintegrasikan ke dalam React component
- Format output harus SVG untuk skalabilitas

### Ukuran:
- Gunakan 24x24px untuk ukuran default
- Pastikan detail terlihat jelas di ukuran kecil
- Hindari detail yang terlalu rumit
- Gunakan bentuk dasar yang mudah dikenali
```

## 💡 Tips Penggunaan

- Gunakan prompt ini untuk menghasilkan icon set secara batch
- Pastikan untuk meminta output dalam format SVG untuk web
- Gunakan warna yang berbeda untuk masing-masing kategori node 
- Simpan setiap icon dengan nama yang sesuai dengan tipe nodenya
- Gunakan sistem grid 24x24 untuk konsistensi
- Buat library icon yang mudah diintegrasikan ke UI framework (React, Vue, dll.)