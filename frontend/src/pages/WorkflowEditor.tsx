import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import axios from 'axios';
import { ArrowLeftIcon, PlayIcon, CloudArrowUpIcon } from '@heroicons/react/24/outline';
import Canvas from '../components/WorkflowCanvas/Canvas';

const WorkflowEditor = () => {
    const { id } = useParams();
    const navigate = useNavigate();
    const isNew = !id;

    const [isRunning, setIsRunning] = useState(false);

    const handleSave = async (nodes: any[], edges: any[]) => {
        const workflowData = {
            id: id || `wf-${Date.now()}`,
            nodes,
            edges,
            updatedAt: new Date().toISOString()
        };
        try {
            const response = await axios.post('/api/v1/workflows', workflowData);
            console.log('Save response:', response.data);
            alert('Workflow saved successfully!');
        } catch (error) {
            console.error('Save failed:', error);
            alert('Failed to save workflow.');
        }
    };

    const [executionInfo, setExecutionInfo] = useState<any>(null);
    const handleRun = async () => {
        setIsRunning(true);
        try {
            const response = await axios.post('/api/v1/workflows/execute', {
                workflow_id: id || 'temp-workflow',
                inputs: {}
            });
            console.log('Execution started:', response.data);
            setExecutionInfo(response.data);
            alert(`Workflow execution started! ID: ${response.data.execution_id}`);
        } catch (error) {
            console.error('Failed to run workflow:', error);
            alert('Failed to run workflow. Ensure backend is running.');
        } finally {
            setIsRunning(false);
        }
    };

    return (
        <div className="h-screen flex flex-col bg-slate-900">
            {/* Header */}
            <div className="h-16 bg-slate-800 border-b border-slate-700 flex items-center justify-between px-6">
                <div className="flex items-center space-x-4">
                    <button
                        onClick={() => navigate('/workflows')}
                        className="p-2 hover:bg-slate-700 rounded-lg text-gray-400 hover:text-white transition-colors"
                    >
                        <ArrowLeftIcon className="h-5 w-5" />
                    </button>
                    <div>
                        <h1 className="text-lg font-bold text-white">
                            {isNew ? 'New Workflow' : 'Edit Workflow'}
                        </h1>
                        <p className="text-xs text-gray-400">
                            {isNew ? 'Unsaved' : `ID: ${id}`}
                        </p>
                    </div>
                </div>

                <div className="flex items-center space-x-3">
                    <button
                        onClick={handleRun}
                        className="flex items-center space-x-2 px-4 py-2 bg-slate-700 text-white rounded-lg hover:bg-slate-600 transition-colors"
                    >
                        <PlayIcon className="h-4 w-4" />
                        <span>{isRunning ? 'Running...' : 'Run'}</span>
                    </button>
                    <button
                        onClick={() => handleSave([], [])} // Canvas will handle actual data
                        className="flex items-center space-x-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
                    >
                        <CloudArrowUpIcon className="h-4 w-4" />
                        <span>Save</span>
                    </button>
                </div>
            </div>

            {/* Canvas Area */}
            <div className="flex-1 overflow-hidden">
                <Canvas onSave={handleSave} onRun={handleRun} />
            </div>
            {/* Execution Result */}
            {executionInfo && (
                <div className="p-4 bg-slate-800 border-t border-slate-700 text-gray-200">
                    <h3 className="text-lg font-semibold mb-2">Execution Result</h3>
                    <pre className="whitespace-pre-wrap text-sm">
                        {JSON.stringify(executionInfo, null, 2)}
                    </pre>
                </div>
            )}
        </div>
    );
};

export default WorkflowEditor;
