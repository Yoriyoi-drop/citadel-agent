# Auto-Save & LocalStorage Persistence

## Overview
Citadel Agent now automatically saves your workflows to browser localStorage, preventing data loss when you refresh the page or accidentally close the browser.

## Features

### ✅ Auto-Save
- **Automatic**: Every change is saved instantly
- **No manual save needed**: Just work on your workflow
- **Real-time**: Updates happen immediately

### ✅ Persistence
- **Survives refresh**: F5 won't delete your work
- **Survives browser close**: Come back anytime
- **Multiple workflows**: All workflows are saved
- **Current state**: Remembers which workflow you were editing

## What Gets Saved

### Saved to LocalStorage:
- ✅ All workflows (list)
- ✅ Current workflow being edited
- ✅ All nodes and their positions
- ✅ All connections/edges
- ✅ Node configurations
- ✅ Workflow settings

### NOT Saved (Session Only):
- ❌ Selected nodes/edges
- ❌ Loading states
- ❌ Error messages
- ❌ UI state (zoom, pan)

## How It Works

### Technical Implementation:
```typescript
// Zustand store with persist middleware
export const useWorkflowStore = create<WorkflowState>()(
  persist(
    subscribeWithSelector((set, get) => ({
      // ... store logic
    })),
    {
      name: 'citadel-workflow-storage', // localStorage key
      partialize: (state) => ({
        currentWorkflow: state.currentWorkflow,
        workflows: state.workflows,
      }),
    }
  )
);
```

### Storage Key:
- **Key**: `citadel-workflow-storage`
- **Location**: Browser localStorage
- **Format**: JSON

## Usage

### For Users:
1. **Create workflow**: Just start adding nodes
2. **Refresh anytime**: Your work is safe
3. **Close browser**: Come back later, it's still there
4. **No save button**: Everything is automatic

### For Developers:
```typescript
// Access the store
const { currentWorkflow, workflows, addNode } = useWorkflowStore();

// Add a node (auto-saves)
addNode(newNode);

// Update workflow (auto-saves)
updateWorkflow(id, { name: 'New Name' });

// Delete workflow (auto-saves)
deleteWorkflow(id);
```

## Storage Limits

### Browser Limits:
- **localStorage**: ~5-10MB per domain
- **Typical workflow**: ~10-50KB
- **Estimated capacity**: 100-500 workflows

### Best Practices:
- ✅ Keep workflows focused
- ✅ Delete unused workflows
- ✅ Export important workflows as JSON
- ✅ Clear old workflows periodically

## Clearing Data

### Manual Clear:
```javascript
// Open browser console (F12)
localStorage.removeItem('citadel-workflow-storage');
// Then refresh page
```

### Clear All:
```javascript
// Open browser console (F12)
localStorage.clear();
// Then refresh page
```

### From UI (Future Feature):
- Settings → Clear All Data
- Workflow → Delete Workflow

## Data Export/Import

### Export Workflow (Manual):
```javascript
// Get workflow from store
const workflow = useWorkflowStore.getState().currentWorkflow;

// Download as JSON
const json = JSON.stringify(workflow, null, 2);
const blob = new Blob([json], { type: 'application/json' });
const url = URL.createObjectURL(blob);
const a = document.createElement('a');
a.href = url;
a.download = `workflow-${workflow.id}.json`;
a.click();
```

### Import Workflow (Manual):
```javascript
// Read JSON file
const workflow = JSON.parse(fileContent);

// Add to store
useWorkflowStore.getState().addWorkflow(workflow);
```

## Troubleshooting

### Workflow Not Saving?
1. Check browser console for errors
2. Check localStorage quota
3. Try incognito mode (fresh start)
4. Clear old data

### Workflow Lost After Refresh?
1. Check if localStorage is enabled
2. Check browser privacy settings
3. Check if in incognito mode
4. Try different browser

### Storage Full?
1. Delete unused workflows
2. Export important workflows
3. Clear localStorage
4. Use smaller workflows

## Privacy & Security

### Data Location:
- **Local only**: Never sent to server
- **Browser only**: Stays on your device
- **No cloud**: No external storage

### Security:
- ⚠️ **Not encrypted**: Anyone with access to your browser can see it
- ⚠️ **Not backed up**: Clear browser data = lose workflows
- ⚠️ **Not synced**: Different browsers = different data

### Recommendations:
- ✅ Export important workflows regularly
- ✅ Use browser password protection
- ✅ Don't store sensitive data in workflows
- ✅ Backup exported JSON files

## Future Enhancements

### Planned Features:
- [ ] Cloud sync (optional)
- [ ] Automatic backups
- [ ] Version history
- [ ] Conflict resolution
- [ ] Collaborative editing
- [ ] Encrypted storage

## Related Files
- `frontend/src/stores/workflowStore.ts` - Store implementation
- `frontend/package.json` - Zustand dependency

## Testing

### Test Auto-Save:
1. Create a workflow
2. Add some nodes
3. Refresh page (F5)
4. ✅ Nodes should still be there

### Test Multiple Workflows:
1. Create workflow A
2. Create workflow B
3. Switch between them
4. Refresh page
5. ✅ Both should be saved

### Test Data Persistence:
1. Create workflow
2. Close browser completely
3. Open browser again
4. Navigate to app
5. ✅ Workflow should be there

## Status
✅ **Implemented and Working**
- Auto-save on every change
- Persists through refresh
- Supports multiple workflows
- No data loss on refresh
