import { NextRequest, NextResponse } from 'next/server';

// Mock workflow data - in real app, this would come from a database
let workflows = [
  {
    id: '1',
    name: 'Customer Data Processing',
    description: 'Process customer data from CRM to database',
    nodes: [],
    edges: [],
    settings: {
      autoSave: true,
      errorHandling: 'stop' as const,
      retryCount: 3
    },
    createdAt: new Date('2024-01-10'),
    updatedAt: new Date('2024-01-15'),
    version: 1,
    isActive: true
  },
  {
    id: '2',
    name: 'Email Campaign Automation',
    description: 'Send automated emails to subscribers',
    nodes: [],
    edges: [],
    settings: {
      autoSave: true,
      errorHandling: 'continue' as const,
      retryCount: 2
    },
    createdAt: new Date('2024-01-08'),
    updatedAt: new Date('2024-01-14'),
    version: 3,
    isActive: true
  }
];

export async function GET() {
  return NextResponse.json({
    success: true,
    data: workflows
  });
}

export async function POST(request: NextRequest) {
  try {
    const workflowData = await request.json();
    
    const newWorkflow = {
      id: Date.now().toString(),
      ...workflowData,
      createdAt: new Date(),
      updatedAt: new Date(),
      version: 1,
      isActive: true
    };
    
    workflows.push(newWorkflow);
    
    return NextResponse.json({
      success: true,
      data: newWorkflow
    });
  } catch (error) {
    return NextResponse.json(
      { error: 'Failed to create workflow' },
      { status: 500 }
    );
  }
}