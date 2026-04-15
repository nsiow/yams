package overlay

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	json "github.com/bytedance/sonic"
	"github.com/nsiow/yams/pkg/entities"
)

const (
	attrID        = "id"
	attrName      = "name"
	attrCreatedAt = "createdAt"
	attrData      = "data"
	nameIndexName = "name-index"
)

// DynamoDBClient defines the DynamoDB operations used by the store.
// This interface enables testing with mocks.
type DynamoDBClient interface {
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
	Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
	Scan(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
}

// DynamoDBAdminClient extends DynamoDBClient with table management operations.
type DynamoDBAdminClient interface {
	DynamoDBClient
	DescribeTable(ctx context.Context, params *dynamodb.DescribeTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DescribeTableOutput, error)
	CreateTable(ctx context.Context, params *dynamodb.CreateTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.CreateTableOutput, error)
}

// DynamoDBStore is a DynamoDB-backed implementation of the Store interface.
type DynamoDBStore struct {
	client    DynamoDBClient
	tableName string
}

// NewDynamoDBStore creates a new DynamoDB-backed overlay store.
// If the table doesn't exist, it will be created with the required schema.
func NewDynamoDBStore(tableName string) (*DynamoDBStore, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := dynamodb.NewFromConfig(cfg)

	// Ensure table exists
	if err := ensureTableExists(context.Background(), client, tableName); err != nil {
		return nil, err
	}

	return &DynamoDBStore{
		client:    client,
		tableName: tableName,
	}, nil
}

// NewDynamoDBStoreWithClient creates a DynamoDB store with a custom client.
// Useful for testing with mocks. Does not auto-create the table.
func NewDynamoDBStoreWithClient(client DynamoDBClient, tableName string) *DynamoDBStore {
	return &DynamoDBStore{
		client:    client,
		tableName: tableName,
	}
}

// ensureTableExists checks if the table exists and creates it if not.
func ensureTableExists(ctx context.Context, client DynamoDBAdminClient, tableName string) error {
	// Check if table exists
	_, err := client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: &tableName,
	})
	if err == nil {
		slog.Info("dynamodb table exists", "table", tableName)
		return nil
	}

	// Check if error is "table not found"
	var notFound *types.ResourceNotFoundException
	if !errors.As(err, &notFound) {
		return fmt.Errorf("failed to describe table: %w", err)
	}

	// Create table
	slog.Info("creating dynamodb table", "table", tableName)
	_, err = client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: &tableName,
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String(attrID), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String(attrName), AttributeType: types.ScalarAttributeTypeS},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String(attrID), KeyType: types.KeyTypeHash},
		},
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String(nameIndexName),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String(attrName), KeyType: types.KeyTypeHash},
				},
				Projection: &types.Projection{ProjectionType: types.ProjectionTypeAll},
			},
		},
		BillingMode: types.BillingModePayPerRequest,
	})
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	// Wait for table to be active
	slog.Info("waiting for table to become active", "table", tableName)
	return waitForTableActive(ctx, client, tableName)
}

// waitForTableActive polls until the table status is ACTIVE.
func waitForTableActive(ctx context.Context, client DynamoDBAdminClient, tableName string) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for i := 0; i < 60; i++ {
		result, err := client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
			TableName: &tableName,
		})
		if err != nil {
			return fmt.Errorf("failed to describe table while waiting: %w", err)
		}

		if result.Table.TableStatus == types.TableStatusActive {
			slog.Info("dynamodb table is active", "table", tableName)
			return nil
		}

		slog.Debug("waiting for table", "table", tableName, "status", result.Table.TableStatus)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}

	return fmt.Errorf("timeout waiting for table %s to become active", tableName)
}

// Create stores a new overlay. Returns ErrAlreadyExists if an overlay
// with the same ID already exists.
func (s *DynamoDBStore) Create(ctx context.Context, overlay *entities.Overlay) error {
	err := s.putOverlay(ctx, overlay, "attribute_not_exists(id)")
	if isConditionalCheckFailed(err) {
		return ErrAlreadyExists
	}
	return err
}

// Get retrieves an overlay by ID. Returns ErrNotFound if not found.
func (s *DynamoDBStore) Get(ctx context.Context, id string) (*entities.Overlay, error) {
	result, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &s.tableName,
		Key:       s.primaryKey(id),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get overlay: %w", err)
	}
	if result.Item == nil {
		return nil, ErrNotFound
	}
	return s.itemToOverlay(result.Item)
}

// GetByName retrieves an overlay by name using the GSI. Returns ErrNotFound if not found.
func (s *DynamoDBStore) GetByName(ctx context.Context, name string) (*entities.Overlay, error) {
	result, err := s.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              &s.tableName,
		IndexName:              aws.String(nameIndexName),
		KeyConditionExpression: aws.String("#n = :name"),
		ExpressionAttributeNames: map[string]string{
			"#n": attrName,
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":name": &types.AttributeValueMemberS{Value: name},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query overlay by name: %w", err)
	}
	if len(result.Items) == 0 {
		return nil, ErrNotFound
	}
	return s.itemToOverlay(result.Items[0])
}

// Update replaces an existing overlay. Returns ErrNotFound if not found.
func (s *DynamoDBStore) Update(ctx context.Context, overlay *entities.Overlay) error {
	err := s.putOverlay(ctx, overlay, "attribute_exists(id)")
	if isConditionalCheckFailed(err) {
		return ErrNotFound
	}
	return err
}

// Delete removes an overlay by ID. Returns ErrNotFound if not found.
func (s *DynamoDBStore) Delete(ctx context.Context, id string) error {
	_, err := s.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName:           &s.tableName,
		Key:                 s.primaryKey(id),
		ConditionExpression: aws.String("attribute_exists(id)"),
	})
	if isConditionalCheckFailed(err) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("failed to delete overlay: %w", err)
	}
	return nil
}

// List returns summaries of all overlays, optionally filtered by a search query.
func (s *DynamoDBStore) List(ctx context.Context, query string) ([]entities.OverlaySummary, error) {
	query = strings.ToLower(query)
	var summaries []entities.OverlaySummary
	var lastKey map[string]types.AttributeValue

	for {
		result, err := s.client.Scan(ctx, &dynamodb.ScanInput{
			TableName:         &s.tableName,
			ExclusiveStartKey: lastKey,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to scan overlays: %w", err)
		}

		for _, item := range result.Items {
			overlay, err := s.itemToOverlay(item)
			if err != nil {
				slog.Warn("failed to parse overlay item", "error", err)
				continue
			}

			// Filter by query if provided
			if query != "" && !strings.Contains(strings.ToLower(overlay.Name), query) {
				continue
			}

			summaries = append(summaries, overlay.Summary())
		}

		lastKey = result.LastEvaluatedKey
		if lastKey == nil {
			break
		}
	}

	return summaries, nil
}

// Exists checks if an overlay with the given ID exists.
func (s *DynamoDBStore) Exists(ctx context.Context, id string) (bool, error) {
	result, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName:            &s.tableName,
		Key:                  s.primaryKey(id),
		ProjectionExpression: aws.String(attrID),
	})
	if err != nil {
		return false, fmt.Errorf("failed to check overlay existence: %w", err)
	}
	return result.Item != nil, nil
}

// primaryKey returns the DynamoDB key for an overlay ID.
func (s *DynamoDBStore) primaryKey(id string) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		attrID: &types.AttributeValueMemberS{Value: id},
	}
}

// putOverlay marshals and stores an overlay with the given condition expression.
func (s *DynamoDBStore) putOverlay(ctx context.Context, overlay *entities.Overlay, condition string) error {
	dataJSON, err := json.Marshal(overlay.ToData())
	if err != nil {
		return fmt.Errorf("failed to marshal overlay: %w", err)
	}

	_, err = s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &s.tableName,
		Item: map[string]types.AttributeValue{
			attrID:        &types.AttributeValueMemberS{Value: overlay.ID},
			attrName:      &types.AttributeValueMemberS{Value: overlay.Name},
			attrCreatedAt: &types.AttributeValueMemberS{Value: overlay.CreatedAt.Format(time.RFC3339Nano)},
			attrData:      &types.AttributeValueMemberS{Value: string(dataJSON)},
		},
		ConditionExpression: aws.String(condition),
	})
	if err != nil && !isConditionalCheckFailed(err) {
		return fmt.Errorf("failed to put overlay: %w", err)
	}
	return err
}

// itemToOverlay converts a DynamoDB item to an Overlay.
func (s *DynamoDBStore) itemToOverlay(item map[string]types.AttributeValue) (*entities.Overlay, error) {
	dataVal, exists := item[attrData]
	if !exists || dataVal == nil {
		return nil, fmt.Errorf("missing data attribute")
	}
	dataAttr, ok := dataVal.(*types.AttributeValueMemberS)
	if !ok {
		return nil, fmt.Errorf("invalid data attribute type")
	}

	var data entities.OverlayData
	if err := json.Unmarshal([]byte(dataAttr.Value), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal overlay data: %w", err)
	}

	return entities.FromData(data), nil
}

// isConditionalCheckFailed checks if the error is a conditional check failure.
func isConditionalCheckFailed(err error) bool {
	var ccf *types.ConditionalCheckFailedException
	return errors.As(err, &ccf)
}
