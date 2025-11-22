package plugins

import (
	"context"
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

// NodePlugin defines the interface for node plugins
type NodePlugin interface {
	Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error)
	GetConfigSchema() map[string]interface{}
	GetMetadata() NodeMetadata
}

// NodeMetadata contains metadata about the node plugin
type NodeMetadata struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Author      string                 `json:"author"`
	Category    string                 `json:"category"` // ai, security, database, etc.
	Schema      map[string]interface{} `json:"schema"`   // JSON schema for configuration
}

// NodePluginRPCServer is the RPC server implementation
type NodePluginRPCServer struct {
	Impl NodePlugin
}

// ExecuteArgs holds the arguments for Execute method
type ExecuteArgs struct {
	Inputs map[string]interface{} `json:"inputs"`
}

// ExecuteReply holds the reply for Execute method
type ExecuteReply struct {
	Outputs map[string]interface{} `json:"outputs"`
	Error   string                 `json:"error"`
}

// GetConfigSchemaReply holds the reply for GetConfigSchema method
type GetConfigSchemaReply struct {
	Schema map[string]interface{} `json:"schema"`
}

// GetMetadataReply holds the reply for GetMetadata method
type GetMetadataReply struct {
	Metadata NodeMetadata `json:"metadata"`
}

func (s *NodePluginRPCServer) Execute(args *ExecuteArgs, reply *ExecuteReply) error {
	outputs, err := s.Impl.Execute(context.Background(), args.Inputs)
	if err != nil {
		reply.Error = err.Error()
	} else {
		reply.Outputs = outputs
	}
	return nil
}

func (s *NodePluginRPCServer) GetConfigSchema(args interface{}, reply *GetConfigSchemaReply) error {
	reply.Schema = s.Impl.GetConfigSchema()
	return nil
}

func (s *NodePluginRPCServer) GetMetadata(args interface{}, reply *GetMetadataReply) error {
	reply.Metadata = s.Impl.GetMetadata()
	return nil
}

// NodePluginRPCClient is the RPC client implementation
type NodePluginRPCClient struct {
	client *rpc.Client
}

func (g *NodePluginRPCClient) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	args := &ExecuteArgs{Inputs: inputs}
	reply := &ExecuteReply{}

	err := g.client.Call("Plugin.Execute", args, reply)
	if err != nil {
		return nil, err
	}

	if reply.Error != "" {
		return nil, &PluginError{reply.Error}
	}

	return reply.Outputs, nil
}

func (g *NodePluginRPCClient) GetConfigSchema() map[string]interface{} {
	reply := &GetConfigSchemaReply{}
	err := g.client.Call("Plugin.GetConfigSchema", struct{}{}, reply)
	if err != nil {
		return nil
	}
	return reply.Schema
}

func (g *NodePluginRPCClient) GetMetadata() NodeMetadata {
	reply := &GetMetadataReply{}
	err := g.client.Call("Plugin.GetMetadata", struct{}{}, reply)
	if err != nil {
		return NodeMetadata{}
	}
	return reply.Metadata
}

// PluginError is a custom error type for plugin errors
type PluginError struct {
	Message string
}

func (e *PluginError) Error() string {
	return e.Message
}

// NodePluginImpl is the implementation of plugin.Plugin so we can serve/consume this
type NodePluginImpl struct {
	Impl NodePlugin
}

func (p *NodePluginImpl) Server(*plugin.MuxBroker) (interface{}, error) {
	return &NodePluginRPCServer{Impl: p.Impl}, nil
}

func (NodePluginImpl) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &NodePluginRPCClient{client: c}, nil
}

// Handshake is the handshake protocol used by go-plugin
var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "CITADEL_PLUGIN",
	MagicCookieValue: "citadel_agent",
}