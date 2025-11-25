import React from 'react';
import { NodePalette } from './NodePalette';

export function WorkflowBuilder() {
    return (
        <div className="flex h-full">
            {/* Node palette on the left */}
            <div className="w-64 border-r bg-muted/20">
                <NodePalette />
            </div>
            {/* Main canvas area */}
            <div className="flex-1 flex items-center justify-center">
                <p className="text-muted-foreground">Workflow canvas placeholder</p>
            </div>
        </div>
    );
}

export default WorkflowBuilder;
