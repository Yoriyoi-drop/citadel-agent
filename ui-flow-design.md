# Citadel Agent - Visual Flow Builder

Desain UI untuk visual flow builder yang memungkinkan pengguna membuat workflow secara visual.

## Komponen Utama

### 1. Node Library Panel
Panel sebelah kiri yang berisi berbagai jenis node yang dapat digunakan dalam workflow.

### 2. Canvas / Flow Editor 
Area kerja utama tempat pengguna menyusun dan menghubungkan node-node.

### 3. Inspector Panel
Panel sebelah kanan untuk mengatur konfigurasi node yang dipilih.

### 4. Execution Logs
Panel bawah untuk melihat log eksekusi secara real-time.

## Wireframe UI

```
┌─────────────────────────────────────────────────────────────────┐
│  CITADEL AGENT - WORKFLOW STUDIO                     [ (_) [] X] │
├─────────────────┬─────────────────────────────────────────────────┤
│                 │    ┌─ DRAG NODE TO CANVAS ──────────────────┐    │
│  NODE LIBRARY   │    │                                      │    │
│                 │    │  ┌─────────────────────────────────┐  │    │
│  [✓] Trigger   │    │  │                                 │  │    │
│     • Cron      │    │  │            CANVAS               │  │    │
│     • Webhook   │    │  │                                 │  │    │
│     • Event     │    │  │                                 │  │    │
│                 │    │  │                                 │  │    │
│  [✓] Actions   │    │  │                                 │  │    │
│     • HTTP      │    │  │                                 │  │    │
│     • Database  │    │  │                                 │  │    │
│     • Script    │    │  │                                 │  │    │
│     • AI Agent  │    │  │                                 │  │    │
│                 │    │  └─────────────────────────────────┘  │    │
│  [✓] Logic     │    │                                    [ ] │    │
│     • Condition │    │                                      │    │
│     • Loop      │    └─────────────────────────────────────────┘    │
│     • Delay     │                                                   │
│                 │    INSPECTOR PANEL      EXECUTION LOGS           │
│  [✓] Output    │    ┌─────────────────┐  ┌──────────────────────┐ │
│     • Variable  │    │                 │  │ [●] 15:32:45 Start │ │
│     • Return    │    │ NODE SETTINGS   │  │ [●] 15:32:46 HTTP  │ │
│                 │    │                 │  │ [●] 15:32:47 Cond  │ │
│                 │    │                 │  │ [●] 15:32:48 End   │ │
│                 │    │                 │  └──────────────────────┘ │
│                 │    └─────────────────┘                          │
└─────────────────┴─────────────────────────────────────────────────┘
```

## Desain Detail Setiap Komponen

### 1. Node Library Panel
```
┌─────────────────────────┐
│      NODE LIBRARY       │
├─────────────────────────┤
│ 🔔 TRIGGERS             │
│ [  ] Cron Schedule      │
│ [  ] Webhook            │
│ [  ] Manual Trigger     │
│ [  ] Event Listener     │
│                         │
│ ⚡ ACTIONS              │
│ [  ] HTTP Request       │
│ [  ] Database Query     │
│ [  ] Execute Script     │
│ [  ] Send Email         │
│ [  ] AI Agent Execute   │
│                         │
│ 🧠 LOGIC                │
│ [  ] Condition          │
│ [  ] Loop               │
│ [  ] Switch Case        │
│ [  ] Delay              │
│                         │
│ ➡️ OUTPUTS              │
│ [  ] Set Variable       │
│ [  ] Return Value       │
│ [  ] Save to File       │
│ [  ] Trigger Next WF    │
└─────────────────────────┘
```

### 2. Canvas / Flow Editor
- Drag & Drop interface
- Node-to-node connection lines
- Zoom and pan functionality
- Grid alignment
- Keyboard shortcuts

### 3. Inspector Panel (Contoh)
```
┌─────────────────────────┐
│    NODE CONFIGURATION   │
├─────────────────────────┤
│ HTTP REQUEST            │
│                         │
│ Method: [GET ▼]         │
│ URL: [https://api.exampl│
│ Headers:                │
│ [+] Add Header          │
│                         │
│ Body:                   │
│ {                       │
│   "key": "value"        │
│ }                       │
│                         │
│ Timeout: [30] seconds   │
│ Retry: [3] times        │
│                         │
│ ┌─────────────────────┐ │
│ │      TEST           │ │
│ └─────────────────────┘ │
└─────────────────────────┘
```

## Component Specifications

### Core Components
1. **FlowCanvas** - Area utama untuk menyusun workflow
2. **NodeComponent** - Representasi visual dari masing-masing node
3. **ConnectorLine** - Garis yang menghubungkan antar node
4. **NodeLibrary** - Panel untuk memilih jenis node
5. **NodeInspector** - Panel untuk mengkonfigurasi node
6. **ExecutionPanel** - Panel untuk melihat log eksekusi

### UI/UX Principles
- Drag and drop intuitif
- Visual feedback saat menyambungkan node
- Undo/redo functionality
- Save/load workflow
- Validation errors clearly displayed
- Responsive design

## Teknologi yang Disarankan
- **React** dengan **React Flow** untuk canvas
- **Redux** atau **Zustand** untuk state management
- **Tailwind CSS** untuk styling
- **TypeScript** untuk type safety

## Contoh Workflow JSON
```json
{
  "id": "wf_http_to_condition",
  "name": "HTTP to Condition Example",
  "description": "Example workflow with HTTP request followed by condition",
  "nodes": [
    {
      "id": "trigger_1",
      "type": "webhook_trigger",
      "position": {"x": 0, "y": 100},
      "config": {
        "path": "/api/webhook",
        "methods": ["POST"]
      }
    },
    {
      "id": "http_1",
      "type": "http_request",
      "position": {"x": 250, "y": 100},
      "config": {
        "method": "GET",
        "url": "https://api.example.com/data",
        "headers": {},
        "timeout": 30
      }
    },
    {
      "id": "cond_1",
      "type": "condition",
      "position": {"x": 500, "y": 100},
      "config": {
        "expression": "{{http_1.response.status}} === 200"
      }
    },
    {
      "id": "success_1",
      "type": "return_value",
      "position": {"x": 750, "y": 50},
      "config": {
        "value": "{{http_1.response.data}}"
      }
    },
    {
      "id": "fail_1",
      "type": "return_value",
      "position": {"x": 750, "y": 150},
      "config": {
        "value": {"error": "Request failed"}
      }
    }
  ],
  "connections": [
    {
      "source": "trigger_1",
      "target": "http_1"
    },
    {
      "source": "http_1",
      "target": "cond_1"
    },
    {
      "source": "cond_1.success",
      "target": "success_1"
    },
    {
      "source": "cond_1.failure",
      "target": "fail_1"
    }
  ]
}
```

## Implementation Steps

### Phase 1: Basic Layout
- Setup React project with React Flow
- Create basic panels (library, canvas, inspector)
- Implement draggable nodes

### Phase 2: Node Implementation
- Create base node component
- Implement different node types
- Add connection functionality

### Phase 3: Configuration
- Implement inspector panel
- Add configuration forms for each node type
- Add validation

### Phase 4: Execution Integration
- Connect to Citadel Agent backend
- Add execution logs panel
- Add run/debug capabilities

### Phase 5: Advanced Features
- Add undo/redo functionality
- Add copy/paste nodes
- Add import/export workflow
- Add collaboration features

## Styling Guidelines

### Color Palette
- Primary: #4F46E5 (Indigo - untuk item terpilih)
- Secondary: #6B7280 (Gray - untuk elemen UI)
- Success: #10B981 (Green - untuk status sukses)
- Warning: #F59E0B (Amber - untuk peringatan)
- Error: #EF4444 (Red - untuk error)
- Background: #FFFFFF (Putih) atau #F9FAFB (Abu terang)

### Typography
- Main headings: Inter/Sans-serif, bold
- Body text: 14px, regular weight
- Code snippets: Monospace, 13px

### Spacing
- Consistent 8px grid system
- Adequate white space between elements
- Responsive design principles

## Interaction Patterns

### Drag and Drop
- Visual feedback when dragging nodes
- Snap to grid when placing
- Connection hints when close to valid target

### Context Menus
- Right-click for node options
- Quick actions menu
- Delete/clone/duplicate options

### Keyboard Shortcuts
- Ctrl+C/V/X untuk copy/paste/cut
- Ctrl+Z/Y untuk undo/redo
- Del untuk delete selected
- Ctrl+D untuk duplicate

## Performance Considerations

- Virtualized rendering for large workflows
- Efficient state updates
- Connection line optimizations
- Lazy loading of node configurations

## Accessibility Features

- Keyboard navigation
- Screen reader compatibility
- High contrast mode
- Focus indicators
- ARIA labels for all interactive elements

## Mobile Responsiveness

While the primary interface is desktop-focused, consider:
- Collapsible panels on smaller screens
- Touch-friendly controls
- Simplified view for mobile devices

## Real-time Collaboration Features

- Live cursor positions
- Concurrent editing
- Conflict resolution
- Change history
- User presence indicators

## Testing Strategy

- Unit tests for node components
- Integration tests for canvas interactions
- End-to-end tests for user workflows
- Performance tests for large workflows
- Accessibility tests