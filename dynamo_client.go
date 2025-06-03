package gogi

import (
	"context"
	"fmt"

	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	appconfig "github.com/dejaniskra/go-gi/internal/config"
)

type DynamoClient struct {
	Client *dynamodb.Client
}

var (
	dynamoClients = make(map[string]*DynamoClient)
	dynamoMu      sync.Mutex
)

func GetDynamoClient(role string) (*DynamoClient, error) {
	dynamoMu.Lock()
	defer dynamoMu.Unlock()

	if client, ok := dynamoClients[role]; ok {
		return client, nil
	}

	cfg := appconfig.GetConfig().Dynamo[role]
	if cfg == nil {
		return nil, fmt.Errorf("no Dynamo config found for role: %s", role)
	}

	client, err := newDynamoClient(cfg)
	if err != nil {
		return nil, err
	}

	dynamoClients[role] = client
	return client, nil
}

func newDynamoClient(cfg *appconfig.DynamoConfig) (*DynamoClient, error) {
	var awsCfg aws.Config
	var err error

	if cfg.AccessKey != "" && cfg.SecretKey != "" {
		creds := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""))
		awsCfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(cfg.Region),
			config.WithCredentialsProvider(creds),
		)
	} else {
		awsCfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(cfg.Region),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	if cfg.Endpoint != nil {
		awsCfg.EndpointResolverWithOptions = aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: *cfg.Endpoint, SigningRegion: cfg.Region}, nil
			},
		)
	}

	client := dynamodb.NewFromConfig(awsCfg)
	GetLogger().Debug(fmt.Sprintf("[DynamoDB] Connected to region=%s endpoint=%v", cfg.Region, cfg.Endpoint))
	return &DynamoClient{Client: client}, nil
}

func (d *DynamoClient) Ping(ctx context.Context) error {
	_, err := d.Client.ListTables(ctx, &dynamodb.ListTablesInput{Limit: aws.Int32(1)})
	return err
}
