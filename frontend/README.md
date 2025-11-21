# Citadel Agent Frontend

The frontend application for Citadel Agent - a powerful workflow automation engine with AI agent capabilities.

## 🚀 Overview

The Citadel Agent frontend provides a modern, intuitive interface for creating and managing automation workflows. Built with React and TypeScript, it features:

- **Visual Workflow Editor**: Drag-and-drop interface for designing workflows
- **Node Library**: Extensive collection of nodes for various integrations
- **Real-time Monitoring**: Track workflow execution and performance
- **AI Agent Integration**: Built-in AI agent capabilities
- **Role-Based Access Control**: Secure access management

## 🛠️ Tech Stack

- **Framework**: React 18
- **Language**: TypeScript
- **Styling**: Tailwind CSS
- **State Management**: Zustand
- **Routing**: React Router v6
- **Workflow Canvas**: React Flow
- **Charts**: Chart.js
- **Drag & Drop**: React DnD

## 📋 Prerequisites

- Node.js 16+ 
- npm or yarn

## 🚀 Getting Started

### 1. Install Dependencies

```bash
npm install
# or
yarn install
```

### 2. Environment Variables

Create a `.env` file in the root of the frontend directory:

```env
REACT_APP_API_URL=http://localhost:5001/api/v1
```

### 3. Development Server

```bash
npm run dev
# or
yarn dev
```

The application will start on `http://localhost:3000`

## 🏗️ Project Structure

```
src/
├── components/          # React components
│   ├── Dashboard/       # Dashboard components
│   ├── Inspector/       # Node inspector panel
│   ├── Sidebar/         # Application sidebar
│   └── WorkflowCanvas/  # Workflow editor canvas
├── services/            # API services
├── store/               # Global state management
├── types/               # TypeScript type definitions
├── hooks/               # Custom React hooks
├── utils/               # Utility functions
├── App.tsx              # Main application component
└── main.tsx             # Application entry point
```

## 🎨 UI Components

### Workflow Canvas
- Interactive node-based workflow editor
- Connection lines between nodes
- Mini-map for navigation
- Real-time node editing

### Sidebar
- Workflow management
- Node library browser
- Navigation menu

### Inspector Panel
- Node property editing
- Parameter configuration
- Data mapping tools
- Execution settings

### Dashboard
- Workflow statistics
- Execution monitoring
- Performance metrics
- Quick actions

## 🔧 Available Scripts

- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm run serve` - Preview production build
- `npm run lint` - Lint code

## 🚢 Deployment

To build the application for production:

```bash
npm run build
```

The build artifacts will be placed in the `dist/` directory.

## 🔐 Security Features

- JWT-based authentication
- Role-based access control
- Secure API communication
- Input validation and sanitization

## 🧪 Testing

Coming soon: Comprehensive test suite with Jest and React Testing Library.

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

Licensed under the Apache 2.0 License.

---

**Citadel Agent Frontend** - Empowering automation through intuitive visual interfaces.