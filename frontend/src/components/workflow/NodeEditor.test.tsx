import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import { NodeEditor } from './NodeEditor';

// Mock the workflow store
jest.mock('@/stores/workflowStore', () => {
    const mockNode = {
        id: 'node-1',
        type: 'http-request',
        data: {
            label: 'Test Node',
            description: 'A test node',
            config: {},
            status: 'idle',
        },
    };
    const mockWorkflow = {
        nodes: [mockNode],
        edges: [],
    };
    return {
        useWorkflowStore: () => ({
            currentWorkflow: mockWorkflow,
            updateNode: jest.fn((id, payload) => {
                // Simple mock: merge payload into node data
                if (id === mockNode.id) {
                    Object.assign(mockNode, payload);
                }
            }),
            selectNodes: jest.fn(),
        }),
    };
});

describe('NodeEditor', () => {
    test('renders node label and description inputs', () => {
        render(<NodeEditor nodeId="node-1" />);
        expect(screen.getByText('Test Node')).toBeInTheDocument();
        const labelInput = screen.getByDisplayValue('Test Node');
        expect(labelInput).toBeInTheDocument();
        const descriptionTextarea = screen.getByDisplayValue('A test node');
        expect(descriptionTextarea).toBeInTheDocument();
    });

    test('executes node and shows output data', async () => {
        // Mock Math.random to avoid error path
        jest.spyOn(Math, 'random').mockReturnValue(0.5);
        render(<NodeEditor nodeId="node-1" />);
        const executeButton = screen.getByRole('button', { name: /execute node/i });
        fireEvent.click(executeButton);
        // Loading state
        expect(executeButton).toBeDisabled();
        await waitFor(() => expect(screen.getByText('Output Data')).toBeInTheDocument(), { timeout: 2000 });
        // Output should appear
        expect(screen.getByText(/execution successful/i)).toBeInTheDocument();
        // Clean up mock
        (Math.random as jest.Mock).mockRestore();
    });

    test('handles execution error', async () => {
        // Force error path
        jest.spyOn(Math, 'random').mockReturnValue(0.1);
        render(<NodeEditor nodeId="node-1" />);
        const executeButton = screen.getByRole('button', { name: /execute node/i });
        fireEvent.click(executeButton);
        await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument(), { timeout: 2000 });
        expect(screen.getByText(/execution failed due to network error/i)).toBeInTheDocument();
        (Math.random as jest.Mock).mockRestore();
    });
});
