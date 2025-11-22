package temporal

import (
	"context"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
)

// Config holds the configuration for Temporal client
type Config struct {
	Address      string
	Namespace    string
	ClientName   string
	ClusterName  string
	Identity     string
}

// TemporalClient wraps the Temporal SDK client with additional functionality
type TemporalClient struct {
	client client.Client
	config *Config
}

// NewTemporalClient creates a new Temporal client
func NewTemporalClient(config *Config) (*TemporalClient, error) {
	opts := client.Options{
		HostPort:  config.Address,
		Namespace: config.Namespace,
	}

	temporalClient, err := client.Dial(opts)
	if err != nil {
		return nil, err
	}

	return &TemporalClient{
		client: temporalClient,
		config: config,
	}, nil
}

// GetClient returns the underlying Temporal client
func (tc *TemporalClient) GetClient() client.Client {
	return tc.client
}

// Close closes the Temporal client connection
func (tc *TemporalClient) Close() {
	tc.client.Close()
}

// ExecuteWorkflow executes a workflow with the given options and parameters
func (tc *TemporalClient) ExecuteWorkflow(ctx context.Context, workflow interface{}, args ...interface{}) (client.WorkflowRun, error) {
	return tc.client.ExecuteWorkflow(ctx, workflow, args...)
}

// GetWorkflow retrieves a workflow run by its ID
func (tc *TemporalClient) GetWorkflow(ctx context.Context, workflowID, runID string) client.WorkflowRun {
	return tc.client.GetWorkflow(ctx, workflowID, runID)
}

// SignalWorkflow sends a signal to a running workflow
func (tc *TemporalClient) SignalWorkflow(ctx context.Context, workflowID, runID, signalName string, arg interface{}) error {
	return tc.client.SignalWorkflow(ctx, workflowID, runID, signalName, arg)
}

// QueryWorkflow queries a running workflow
func (tc *TemporalClient) QueryWorkflow(ctx context.Context, workflowID, runID, queryType string, args ...interface{}) (converter.EncodedValue, error) {
	return tc.client.QueryWorkflow(ctx, workflowID, runID, queryType, args...)
}