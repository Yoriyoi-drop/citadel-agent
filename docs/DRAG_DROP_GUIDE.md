# Drag & Drop Nodes - User Guide

## How to Add Nodes to Workflow

### Method 1: Drag & Drop (Recommended)
1. Open a workflow editor page (`/workflows/[id]` or `/workflows/new`)
2. Find the node you want in the **Node Palette** (left sidebar)
3. Click and hold on the node card
4. Drag it to the canvas area
5. Release to drop the node at the desired position

### Method 2: Click to Add (Alternative)
If drag & drop doesn't work:
1. Click the **Plus (+)** button on the node card
2. Node will be added to the center of the canvas

## Troubleshooting

### Nodes Not Appearing After Drop

**Check Console Logs:**
Open browser DevTools (F12) → Console tab

You should see:
```
Adding node: { id: "http_request_1234567890", type: "http_request", ... }
```

**Common Issues:**

1. **ReactFlow Instance Not Ready**
   - Wait a moment after page loads
   - Console shows: `ReactFlow instance not ready`
   - **Fix**: Refresh the page and try again

2. **No Drag Data Found**
   - Drag gesture was not captured properly
   - Console shows: `No drag data found`
   - **Fix**: Try dragging again, make sure to hold mouse button

3. **No Current Workflow**
   - Should auto-create workflow now
   - If not, navigate to `/workflows/new` first

### Debug Mode

The system now logs all drag & drop operations:
- `console.warn()` for issues
- `console.log()` for successful operations
- `console.error()` for errors

## Technical Details

### Data Transfer Format
```typescript
{
  nodeType: "http_request",
  label: "HTTP Request",
  description: "Make HTTP requests",
  inputs: [...],
  outputs: [...],
  config: {...}
}
```

### Auto-Workflow Creation
If you drop a node without an active workflow:
- System automatically creates a new workflow
- Name: "New Workflow"
- Description: "Untitled workflow"
- You can rename it later in settings

## Features

✅ Drag & drop from palette to canvas  
✅ Auto-create workflow if needed  
✅ Position nodes anywhere on canvas  
✅ Visual feedback during drag  
✅ Error handling with console logs  
✅ Works with all 26+ node types  

## Next Steps

After adding nodes:
1. **Connect nodes**: Drag from output handle to input handle
2. **Configure nodes**: Click node to open editor sidebar
3. **Execute nodes**: Click play button in floating toolbar
4. **Save workflow**: Click save button in header

## Related Files
- `WorkflowBuilder.tsx` - Main canvas component
- `NodePalette.tsx` - Node list sidebar
- `NodeCard.tsx` - Draggable node cards
- `workflowStore.ts` - State management
