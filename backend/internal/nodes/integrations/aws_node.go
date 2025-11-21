// backend/internal/nodes/integrations/aws_node.go
package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	
	"citadel-agent/backend/internal/workflow/core/engine"
)

// AWSServiceType represents the type of AWS service
type AWSServiceType string

const (
	AWSS3          AWSServiceType = "s3"
	AWSEC2         AWSServiceType = "ec2"
	AWSLambda      AWSServiceType = "lambda"
	AWSSNS         AWSServiceType = "sns"
	AWSSQS         AWSServiceType = "sqs"
	AWSDynamoDB    AWSServiceType = "dynamodb"
	AWSSecretsMgmt AWSServiceType = "secretsmanager"
	AWSSSM         AWSServiceType = "ssm"
	AWSIAM         AWSServiceType = "iam"
	AWSRDS         AWSServiceType = "rds"
	AWSCloudWatch  AWSServiceType = "cloudwatch"
)

// AWSOperation represents the operation to perform on AWS service
type AWSOperation string

const (
	// S3 Operations
	S3GetObject     AWSOperation = "s3_get_object"
	S3PutObject     AWSOperation = "s3_put_object"
	S3ListObjects   AWSOperation = "s3_list_objects"
	S3DeleteObject  AWSOperation = "s3_delete_object"

	// EC2 Operations
	EC2DescribeInstances AWSOperation = "ec2_describe_instances"
	EC2StartInstances    AWSOperation = "ec2_start_instances"
	EC2StopInstances     AWSOperation = "ec2_stop_instances"
	EC2TerminateInstances AWSOperation = "ec2_terminate_instances"

	// Lambda Operations
	LambdaInvoke      AWSOperation = "lambda_invoke"
	LambdaList        AWSOperation = "lambda_list_functions"
	LambdaCreate      AWSOperation = "lambda_create_function"
	LambdaUpdate      AWSOperation = "lambda_update_function"

	// SNS Operations
	SNSSendMessage    AWSOperation = "sns_send_message"
	SNSListTopics     AWSOperation = "sns_list_topics"
	SNSSubscribe      AWSOperation = "sns_subscribe"
	SNSUnsubscribe    AWSOperation = "sns_unsubscribe"

	// SQS Operations
	SQSSendMessage    AWSOperation = "sqs_send_message"
	SQSReceiveMessage AWSOperation = "sqs_receive_message"
	SQSListQueues     AWSOperation = "sqs_list_queues"

	// DynamoDB Operations
	DynamoPutItem     AWSOperation = "dynamodb_put_item"
	DynamoGetItem     AWSOperation = "dynamodb_get_item"
	DynamoQuery       AWSOperation = "dynamodb_query"
	DynamoScan        AWSOperation = "dynamodb_scan"
	DynamoUpdateItem  AWSOperation = "dynamodb_update_item"
	DynamoDeleteItem  AWSOperation = "dynamodb_delete_item"

	// Secrets Manager Operations
	SecretsGet        AWSOperation = "secretsmanager_get_secret"
	SecretsCreate     AWSOperation = "secretsmanager_create_secret"
	SecretsUpdate     AWSOperation = "secretsmanager_update_secret"
	SecretsDelete     AWSOperation = "secretsmanager_delete_secret"

	// SSM Operations
	SSMGetParameter   AWSOperation = "ssm_get_parameter"
	SSMPutParameter   AWSOperation = "ssm_put_parameter"
	SSMListParameters AWSOperation = "ssm_list_parameters"

	// IAM Operations
	IAMListUsers      AWSOperation = "iam_list_users"
	IAMListRoles      AWSOperation = "iam_list_roles"
	IAMCreateUser     AWSOperation = "iam_create_user"
	IAMCreateRole     AWSOperation = "iam_create_role"
)

// AWSNodeConfig represents the configuration for an AWS node
type AWSNodeConfig struct {
	ServiceType     AWSServiceType   `json:"service_type"`
	OperationType   AWSOperation     `json:"operation_type"`
	AccessKeyID     string           `json:"access_key_id"`
	SecretAccessKey string           `json:"secret_access_key"`
	SessionToken    string           `json:"session_token,omitempty"`
	Region          string           `json:"region"`
	Endpoint        string           `json:"endpoint,omitempty"` // For custom endpoints
	MaxRetries      int              `json:"max_retries"`
	Timeout         time.Duration    `json:"timeout"`
	EnableTLS       bool             `json:"enable_tls"`
	ProxyURL        string           `json:"proxy_url,omitempty"`
	Params          map[string]interface{} `json:"params"` // Service-specific parameters
}

// AWSNode represents an AWS integration node
type AWSNode struct {
	config *AWSNodeConfig
	cfg    aws.Config
	client interface{} // Will hold the appropriate AWS service client
}

// NewAWSNode creates a new AWS node
func NewAWSNode(config *AWSNodeConfig) (*AWSNode, error) {
	// Set defaults
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.Region == "" {
		config.Region = "us-east-1" // Default region
	}

	// Load AWS config
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(config.Region),
		config.WithSharedCredentials(true),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	node := &AWSNode{
		config: config,
		cfg:    cfg,
	}

	// Initialize the appropriate client based on service type
	if err := node.initializeClient(); err != nil {
		return nil, fmt.Errorf("failed to initialize AWS client: %w", err)
	}

	return node, nil
}

// initializeClient initializes the appropriate AWS service client
func (an *AWSNode) initializeClient() error {
	ctx, cancel := context.WithTimeout(context.Background(), an.config.Timeout)
	defer cancel()

	switch an.config.ServiceType {
	case AWSS3:
		an.client = s3.NewFromConfig(an.cfg)
	case AWSEC2:
		an.client = ec2.NewFromConfig(an.cfg)
	case AWSLambda:
		an.client = lambda.NewFromConfig(an.cfg)
	case AWSSNS:
		an.client = sns.NewFromConfig(an.cfg)
	case AWSSQS:
		an.client = sqs.NewFromConfig(an.cfg)
	case AWSDynamoDB:
		an.client = dynamodb.NewFromConfig(an.cfg)
	case AWSSecretsMgmt:
		an.client = secretsmanager.NewFromConfig(an.cfg)
	case AWSSSM:
		an.client = ssm.NewFromConfig(an.cfg)
	case AWSIAM:
		an.client = iam.NewFromConfig(an.cfg)
	case AWSRDS:
		an.client = rds.NewFromConfig(an.cfg)
	case AWSCloudWatch:
		an.client = cloudwatch.NewFromConfig(an.cfg)
	default:
		return fmt.Errorf("unsupported AWS service type: %s", an.config.ServiceType)
	}

	return nil
}

// Execute executes the AWS operation
func (an *AWSNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Override config values with inputs if provided
	service := an.config.ServiceType
	if svc, exists := inputs["service_type"]; exists {
		if svcStr, ok := svc.(string); ok {
			service = AWSServiceType(svcStr)
		}
	}

	operation := an.config.OperationType
	if op, exists := inputs["operation_type"]; exists {
		if opStr, ok := op.(string); ok {
			operation = AWSOperation(opStr)
		}
	}

	// Merge params from config and inputs
	params := make(map[string]interface{})
	for k, v := range an.config.Params {
		params[k] = v
	}
	for k, v := range inputs {
		if k != "service_type" && k != "operation_type" && k != "access_key_id" && k != "secret_access_key" {
			params[k] = v
		}
	}

	// Prepare context with timeout
	execCtx, cancel := context.WithTimeout(ctx, an.config.Timeout)
	defer cancel()

	// Execute operation based on service and operation type
	switch service {
	case AWSS3:
		return an.executeS3Operation(execCtx, operation, params)
	case AWSEC2:
		return an.executeEC2Operation(execCtx, operation, params)
	case AWSLambda:
		return an.executeLambdaOperation(execCtx, operation, params)
	case AWSSNS:
		return an.executeSNSOperation(execCtx, operation, params)
	case AWSSQS:
		return an.executeSQSOperation(execCtx, operation, params)
	case AWSDynamoDB:
		return an.executeDynamoDBOperation(execCtx, operation, params)
	case AWSSecretsMgmt:
		return an.executeSecretsManagerOperation(execCtx, operation, params)
	case AWSSSM:
		return an.executeSSMOperation(execCtx, operation, params)
	case AWSIAM:
		return an.executeIAMOperation(execCtx, operation, params)
	case AWSRDS:
		return an.executeRDSOperation(execCtx, operation, params)
	case AWSCloudWatch:
		return an.executeCloudWatchOperation(execCtx, operation, params)
	default:
		return nil, fmt.Errorf("unsupported AWS service type: %s", service)
	}
}

// executeS3Operation executes S3 operations
func (an *AWSNode) executeS3Operation(ctx context.Context, operation AWSOperation, params map[string]interface{}) (map[string]interface{}, error) {
	s3Client, ok := an.client.(*s3.Client)
	if !ok {
		return nil, fmt.Errorf("client is not an S3 client")
	}

	switch operation {
	case S3GetObject:
		return an.s3GetObject(ctx, s3Client, params)
	case S3PutObject:
		return an.s3PutObject(ctx, s3Client, params)
	case S3ListObjects:
		return an.s3ListObjects(ctx, s3Client, params)
	case S3DeleteObject:
		return an.s3DeleteObject(ctx, s3Client, params)
	default:
		return nil, fmt.Errorf("unsupported S3 operation: %s", operation)
	}
}

// s3GetObject gets an object from S3
func (an *AWSNode) s3GetObject(ctx context.Context, client *s3.Client, params map[string]interface{}) (map[string]interface{}, error) {
	bucket, exists := params["bucket"]
	if !exists {
		return nil, fmt.Errorf("'bucket' parameter is required for S3 GetObject operation")
	}
	bucketStr, ok := bucket.(string)
	if !ok {
		return nil, fmt.Errorf("'bucket' parameter must be a string")
	}

	key, exists := params["key"]
	if !exists {
		return nil, fmt.Errorf("'key' parameter is required for S3 GetObject operation")
	}
	keyStr, ok := key.(string)
	if !ok {
		return nil, fmt.Errorf("'key' parameter must be a string")
	}

	getObjectInput := &s3.GetObjectInput{
		Bucket: aws.String(bucketStr),
		Key:    aws.String(keyStr),
	}

	// Add optional parameters
	if rangeParam, exists := params["range"]; exists {
		if rangeStr, ok := rangeParam.(string); ok {
			getObjectInput.Range = &rangeStr
		}
	}

	if versionID, exists := params["version_id"]; exists {
		if versionStr, ok := versionID.(string); ok {
			getObjectInput.VersionId = &versionStr
		}
	}

	result, err := client.GetObject(ctx, getObjectInput)

	if err != nil {
		return nil, fmt.Errorf("failed to get S3 object: %w", err)
	}
	defer result.Body.Close()

	// Read the object content
	content := make([]byte, 0)
	if result.Body != nil {
		// In a real implementation, we'd read the full content
		// For this example, we'll just return the metadata
	}

	return map[string]interface{}{
		"success": true,
		"operation": string(S3GetObject),
		"service":   string(AWSS3),
		"bucket":    bucketStr,
		"key":       keyStr,
		"content_length": result.ContentLength,
		"content_type":   aws.ToString(result.ContentType),
		"etag":           aws.ToString(result.ETag),
		"last_modified":  result.LastModified,
		"status_code":    200, // Since the call succeeded at the AWS level
		"timestamp":      time.Now().Unix(),
	}, nil
}

// s3PutObject puts an object to S3
func (an *AWSNode) s3PutObject(ctx context.Context, client *s3.Client, params map[string]interface{}) (map[string]interface{}, error) {
	bucket, exists := params["bucket"]
	if !exists {
		return nil, fmt.Errorf("'bucket' parameter is required for S3 PutObject operation")
	}
	bucketStr, ok := bucket.(string)
	if !ok {
		return nil, fmt.Errorf("'bucket' parameter must be a string")
	}

	key, exists := params["key"]
	if !exists {
		return nil, fmt.Errorf("'key' parameter is required for S3 PutObject operation")
	}
	keyStr, ok := key.(string)
	if !ok {
		return nil, fmt.Errorf("'key' parameter must be a string")
	}

	body, exists := params["body"]
	if !exists {
		return nil, fmt.Errorf("'body' parameter is required for S3 PutObject operation")
	}

	// Convert body to appropriate format
	var bodyReader interface{}
	switch v := body.(type) {
	case string:
		bodyReader = strings.NewReader(v)
	case []byte:
		bodyReader = bytes.NewReader(v)
	default:
		// If it's not a string or bytes, try to marshal to JSON
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("unable to convert body to readable format: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBytes)
	}

	putObjectInput := &s3.PutObjectInput{
		Bucket: aws.String(bucketStr),
		Key:    aws.String(keyStr),
		Body:   bodyReader,
	}

	// Add optional parameters
	if acl, exists := params["acl"]; exists {
		if aclStr, ok := acl.(string); ok {
			putObjectInput.ACL = s3Types.ObjectCannedACLP("public-read") // Example ACL
		}
	}

	if contentType, exists := params["content_type"]; exists {
		if ctStr, ok := contentType.(string); ok {
			putObjectInput.ContentType = &ctStr
		}
	}

	if metadata, exists := params["metadata"]; exists {
		if metaMap, ok := metadata.(map[string]interface{}); ok {
			metadata := make(map[string]string)
			for k, v := range metaMap {
				if vStr, ok := v.(string); ok {
					metadata[k] = vStr
				}
			}
			putObjectInput.Metadata = metadata
		}
	}

	result, err := client.PutObject(ctx, putObjectInput)
	if err != nil {
		return nil, fmt.Errorf("failed to put S3 object: %w", err)
	}

	return map[string]interface{}{
		"success":     true,
		"operation":   string(S3PutObject),
		"service":     string(AWSS3),
		"bucket":      bucketStr,
		"key":         keyStr,
		"etag":        aws.ToString(result.ETag),
		"timestamp":   time.Now().Unix(),
	}, nil
}

// s3ListObjects lists objects in an S3 bucket
func (an *AWSNode) s3ListObjects(ctx context.Context, client *s3.Client, params map[string]interface{}) (map[string]interface{}, error) {
	bucket, exists := params["bucket"]
	if !exists {
		return nil, fmt.Errorf("'bucket' parameter is required for S3 ListObjects operation")
	}
	bucketStr, ok := bucket.(string)
	if !ok {
		return nil, fmt.Errorf("'bucket' parameter must be a string")
	}

	listObjectsInput := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketStr),
	}

	// Add optional parameters
	if prefix, exists := params["prefix"]; exists {
		if prefixStr, ok := prefix.(string); ok {
			listObjectsInput.Prefix = &prefixStr
		}
	}

	if maxKeys, exists := params["max_keys"]; exists {
		if maxKeysFloat, ok := maxKeys.(float64); ok {
			max := int32(maxKeysFloat)
			listObjectsInput.MaxKeys = &max
		}
	}

	if delimiter, exists := params["delimiter"]; exists {
		if delimiterStr, ok := delimiter.(string); ok {
			listObjectsInput.Delimiter = &delimiterStr
		}
	}

	result, err := client.ListObjectsV2(ctx, listObjectsInput)
	if err != nil {
		return nil, fmt.Errorf("failed to list S3 objects: %w", err)
	}

	// Process the response
	objects := make([]map[string]interface{}, len(result.Contents))
	for i, obj := range result.Contents {
		objects[i] = map[string]interface{}{
			"key":          aws.ToString(obj.Key),
			"size":         obj.Size,
			"last_modified": obj.LastModified,
			"etag":         aws.ToString(obj.ETag),
			"storage_class": string(obj.StorageClass),
		}
	}

	return map[string]interface{}{
		"success":     true,
		"operation":   string(S3ListObjects),
		"service":     string(AWSS3),
		"bucket":      bucketStr,
		"objects":     objects,
		"object_count": len(objects),
		"is_truncated": result.IsTruncated,
		"prefix":      aws.ToString(listObjectsInput.Prefix),
		"timestamp":   time.Now().Unix(),
	}, nil
}

// s3DeleteObject deletes an object from S3
func (an *AWSNode) s3DeleteObject(ctx context.Context, client *s3.Client, params map[string]interface{}) (map[string]interface{}, error) {
	bucket, exists := params["bucket"]
	if !exists {
		return nil, fmt.Errorf("'bucket' parameter is required for S3 DeleteObject operation")
	}
	bucketStr, ok := bucket.(string)
	if !ok {
		return nil, fmt.Errorf("'bucket' parameter must be a string")
	}

	key, exists := params["key"]
	if !exists {
		return nil, fmt.Errorf("'key' parameter is required for S3 DeleteObject operation")
	}
	keyStr, ok := key.(string)
	if !ok {
		return nil, fmt.Errorf("'key' parameter must be a string")
	}

	deleteObjectInput := &s3.DeleteObjectInput{
		Bucket: aws.String(bucketStr),
		Key:    aws.String(keyStr),
	}

	if versionID, exists := params["version_id"]; exists {
		if versionStr, ok := versionID.(string); ok {
			deleteObjectInput.VersionId = &versionStr
		}
	}

	result, err := client.DeleteObject(ctx, deleteObjectInput)
	if err != nil {
		return nil, fmt.Errorf("failed to delete S3 object: %w", err)
	}

	return map[string]interface{}{
		"success":     true,
		"operation":   string(S3DeleteObject),
		"service":     string(AWSS3),
		"bucket":      bucketStr,
		"key":         keyStr,
		"delete_marker": result.DeleteMarker,
		"version_id":  aws.ToString(result.VersionId),
		"timestamp":   time.Now().Unix(),
	}, nil
}

// executeEC2Operation executes EC2 operations
func (an *AWSNode) executeEC2Operation(ctx context.Context, operation AWSOperation, params map[string]interface{}) (map[string]interface{}, error) {
	ec2Client, ok := an.client.(*ec2.Client)
	if !ok {
		return nil, fmt.Errorf("client is not an EC2 client")
	}

	switch operation {
	case EC2DescribeInstances:
		return an.ec2DescribeInstances(ctx, ec2Client, params)
	case EC2StartInstances:
		return an.ec2StartInstances(ctx, ec2Client, params)
	case EC2StopInstances:
		return an.ec2StopInstances(ctx, ec2Client, params)
	case EC2TerminateInstances:
		return an.ec2TerminateInstances(ctx, ec2Client, params)
	default:
		return nil, fmt.Errorf("unsupported EC2 operation: %s", operation)
	}
}

// ec2DescribeInstances describes EC2 instances
func (an *AWSNode) ec2DescribeInstances(ctx context.Context, client *ec2.Client, params map[string]interface{}) (map[string]interface{}, error) {
	describeInstancesInput := &ec2.DescribeInstancesInput{}

	// Add optional filters
	var filters []types.Filter
	for key, value := range params {
		if strings.HasPrefix(key, "filter_") {
			filterName := strings.TrimPrefix(key, "filter_")
			if valueStr, ok := value.(string); ok {
				filters = append(filters, types.Filter{
					Name:   aws.String(filterName),
					Values: []string{valueStr},
				})
			}
		}
	}

	if len(filters) > 0 {
		describeInstancesInput.Filters = filters
	}

	// Add instance IDs if specified
	if instanceIDs, exists := params["instance_ids"]; exists {
		if idsSlice, ok := instanceIDs.([]interface{}); ok {
			var ids []string
			for _, id := range idsSlice {
				if idStr, ok := id.(string); ok {
					ids = append(ids, idStr)
				}
			}
			if len(ids) > 0 {
				describeInstancesInput.InstanceIds = ids
			}
		}
	}

	result, err := client.DescribeInstances(ctx, describeInstancesInput)
	if err != nil {
		return nil, fmt.Errorf("failed to describe EC2 instances: %w", err)
	}

	// Process the response
	var instances []map[string]interface{}
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			instanceMap := map[string]interface{}{
				"instance_id":     aws.ToString(instance.InstanceId),
				"instance_type":   string(instance.InstanceType),
				"state":           string(instance.State.Name),
				"image_id":        aws.ToString(instance.ImageId),
				"launch_time":     instance.LaunchTime,
				"public_ip":       aws.ToString(instance.PublicIpAddress),
				"private_ip":      aws.ToString(instance.PrivateIpAddress),
				"subnet_id":       aws.ToString(instance.SubnetId),
				"vpc_id":          aws.ToString(instance.VpcId),
				"architecture":    string(instance.Architecture),
				"platform":        string(instance.PlatformDetails),
				"availability_zone": aws.ToString(instance.Placement.AvailabilityZone),
			}
			
			if instance.Tags != nil {
				tags := make(map[string]string)
				for _, tag := range instance.Tags {
					tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
				}
				instanceMap["tags"] = tags
			}
			
			instances = append(instances, instanceMap)
		}
	}

	return map[string]interface{}{
		"success":      true,
		"operation":    string(EC2DescribeInstances),
		"service":      string(AWSEC2),
		"instances":    instances,
		"instance_count": len(instances),
		"reservation_count": len(result.Reservations),
		"timestamp":    time.Now().Unix(),
	}, nil
}

// ec2StartInstances starts EC2 instances
func (an *AWSNode) ec2StartInstances(ctx context.Context, client *ec2.Client, params map[string]interface{}) (map[string]interface{}, error) {
	instanceIDs, exists := params["instance_ids"]
	if !exists {
		return nil, fmt.Errorf("'instance_ids' parameter is required for EC2 StartInstances operation")
	}

	var ids []string
	if idsSlice, ok := instanceIDs.([]interface{}); ok {
		for _, id := range idsSlice {
			if idStr, ok := id.(string); ok {
				ids = append(ids, idStr)
			}
		}
	} else if idStr, ok := instanceIDs.(string); ok {
		ids = []string{idStr}
	} else {
		return nil, fmt.Errorf("'instance_ids' must be a string or array of strings")
	}

	if len(ids) == 0 {
		return nil, fmt.Errorf("'instance_ids' cannot be empty")
	}

	startInstancesInput := &ec2.StartInstancesInput{
		InstanceIds: ids,
	}

	result, err := client.StartInstances(ctx, startInstancesInput)
	if err != nil {
		return nil, fmt.Errorf("failed to start EC2 instances: %w", err)
	}

	// Process the response
	startingInstances := make([]map[string]interface{}, len(result.StartingInstances))
	for i, instance := range result.StartingInstances {
		startingInstances[i] = map[string]interface{}{
			"instance_id":       aws.ToString(instance.InstanceId),
			"previous_state":    string(instance.PreviousState.Name),
			"current_state":     string(instance.CurrentState.Name),
		}
	}

	return map[string]interface{}{
		"success":         true,
		"operation":       string(EC2StartInstances),
		"service":         string(AWSEC2),
		"instance_ids":    ids,
		"starting_instances": startingInstances,
		"timestamp":       time.Now().Unix(),
	}, nil
}

// RegisterAWSNode registers the AWS node type with the engine
func RegisterAWSNode(registry *engine.NodeRegistry) {
	registry.RegisterNodeType("aws_integration", func(config map[string]interface{}) (engine.NodeInstance, error) {
		var serviceType AWSServiceType
		if svc, exists := config["service_type"]; exists {
			if svcStr, ok := svc.(string); ok {
				serviceType = AWSServiceType(svcStr)
			}
		}

		var operationType AWSOperation
		if op, exists := config["operation_type"]; exists {
			if opStr, ok := op.(string); ok {
				operationType = AWSOperation(opStr)
			}
		}

		var accessKeyID string
		if key, exists := config["access_key_id"]; exists {
			if keyStr, ok := key.(string); ok {
				accessKeyID = keyStr
			}
		}

		var secretAccessKey string
		if secret, exists := config["secret_access_key"]; exists {
			if secretStr, ok := secret.(string); ok {
				secretAccessKey = secretStr
			}
		}

		var sessionToken string
		if token, exists := config["session_token"]; exists {
			if tokenStr, ok := token.(string); ok {
				sessionToken = tokenStr
			}
		}

		var region string
		if reg, exists := config["region"]; exists {
			if regStr, ok := reg.(string); ok {
				region = regStr
			}
		}

		var endpoint string
		if ep, exists := config["endpoint"]; exists {
			if epStr, ok := ep.(string); ok {
				endpoint = epStr
			}
		}

		var maxRetries float64
		if retries, exists := config["max_retries"]; exists {
			if retriesFloat, ok := retries.(float64); ok {
				maxRetries = retriesFloat
			}
		}

		var timeout float64
		if t, exists := config["timeout_seconds"]; exists {
			if tFloat, ok := t.(float64); ok {
				timeout = tFloat
			}
		}

		var enableTLS bool
		if tls, exists := config["enable_tls"]; exists {
			if tlsBool, ok := tls.(bool); ok {
				enableTLS = tlsBool
			}
		}

		var proxyURL string
		if url, exists := config["proxy_url"]; exists {
			if urlStr, ok := url.(string); ok {
				proxyURL = urlStr
			}
		}

		var params map[string]interface{}
		if p, exists := config["params"]; exists {
			if pMap, ok := p.(map[string]interface{}); ok {
				params = pMap
			}
		}

		nodeConfig := &AWSNodeConfig{
			ServiceType:     serviceType,
			OperationType:   operationType,
			AccessKeyID:     accessKeyID,
			SecretAccessKey: secretAccessKey,
			SessionToken:    sessionToken,
			Region:          region,
			Endpoint:        endpoint,
			MaxRetries:      int(maxRetries),
			Timeout:         time.Duration(timeout) * time.Second,
			EnableTLS:       enableTLS,
			ProxyURL:        proxyURL,
			Params:          params,
		}

		return NewAWSNode(nodeConfig)
	})
}