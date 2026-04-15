package overlay

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	json "github.com/bytedance/sonic"
	"github.com/nsiow/yams/pkg/entities"
)

// mockDynamoDBClient implements DynamoDBClient for testing.
type mockDynamoDBClient struct {
	items map[string]map[string]types.AttributeValue // id -> item
}

func newMockClient() *mockDynamoDBClient {
	return &mockDynamoDBClient{
		items: make(map[string]map[string]types.AttributeValue),
	}
}

func (m *mockDynamoDBClient) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	idAttr := params.Item[attrID].(*types.AttributeValueMemberS)
	id := idAttr.Value

	// Check condition expression
	if params.ConditionExpression != nil {
		expr := *params.ConditionExpression
		if expr == "attribute_not_exists(id)" {
			if _, exists := m.items[id]; exists {
				return nil, &types.ConditionalCheckFailedException{Message: aws.String("condition failed")}
			}
		} else if expr == "attribute_exists(id)" {
			if _, exists := m.items[id]; !exists {
				return nil, &types.ConditionalCheckFailedException{Message: aws.String("condition failed")}
			}
		}
	}

	m.items[id] = params.Item
	return &dynamodb.PutItemOutput{}, nil
}

func (m *mockDynamoDBClient) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	idAttr := params.Key[attrID].(*types.AttributeValueMemberS)
	id := idAttr.Value

	item, exists := m.items[id]
	if !exists {
		return &dynamodb.GetItemOutput{Item: nil}, nil
	}

	return &dynamodb.GetItemOutput{Item: item}, nil
}

func (m *mockDynamoDBClient) DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	idAttr := params.Key[attrID].(*types.AttributeValueMemberS)
	id := idAttr.Value

	// Check condition expression
	if params.ConditionExpression != nil && *params.ConditionExpression == "attribute_exists(id)" {
		if _, exists := m.items[id]; !exists {
			return nil, &types.ConditionalCheckFailedException{Message: aws.String("condition failed")}
		}
	}

	delete(m.items, id)
	return &dynamodb.DeleteItemOutput{}, nil
}

func (m *mockDynamoDBClient) Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	// For name-index queries, find by name
	nameValue := params.ExpressionAttributeValues[":name"].(*types.AttributeValueMemberS).Value

	var items []map[string]types.AttributeValue
	for _, item := range m.items {
		nameAttr := item[attrName].(*types.AttributeValueMemberS)
		if nameAttr.Value == nameValue {
			items = append(items, item)
			break
		}
	}

	return &dynamodb.QueryOutput{Items: items}, nil
}

func (m *mockDynamoDBClient) Scan(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
	var items []map[string]types.AttributeValue
	for _, item := range m.items {
		items = append(items, item)
	}
	return &dynamodb.ScanOutput{Items: items}, nil
}

func TestDynamoDBStore_Create(t *testing.T) {
	client := newMockClient()
	store := NewDynamoDBStoreWithClient(client, "test-table")
	ctx := context.Background()

	overlay := entities.NewOverlay("test-overlay")
	overlay.Universe.PutPrincipal(entities.Principal{Arn: "arn:aws:iam::123456789012:role/test"})

	// Create should succeed
	err := store.Create(ctx, overlay)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Create same overlay again should fail
	err = store.Create(ctx, overlay)
	if err != ErrAlreadyExists {
		t.Errorf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestDynamoDBStore_Get(t *testing.T) {
	client := newMockClient()
	store := NewDynamoDBStoreWithClient(client, "test-table")
	ctx := context.Background()

	overlay := entities.NewOverlay("test-overlay")
	overlay.Universe.PutPrincipal(entities.Principal{Arn: "arn:aws:iam::123456789012:role/test"})
	_ = store.Create(ctx, overlay)

	// Get should succeed
	retrieved, err := store.Get(ctx, overlay.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.Name != overlay.Name {
		t.Errorf("expected name %q, got %q", overlay.Name, retrieved.Name)
	}
	if retrieved.ID != overlay.ID {
		t.Errorf("expected ID %q, got %q", overlay.ID, retrieved.ID)
	}
	if retrieved.NumPrincipals() != 1 {
		t.Errorf("expected 1 principal, got %d", retrieved.NumPrincipals())
	}

	// Get non-existent should fail
	_, err = store.Get(ctx, "non-existent")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDynamoDBStore_GetByName(t *testing.T) {
	client := newMockClient()
	store := NewDynamoDBStoreWithClient(client, "test-table")
	ctx := context.Background()

	overlay := entities.NewOverlay("test-overlay")
	_ = store.Create(ctx, overlay)

	// GetByName should succeed
	retrieved, err := store.GetByName(ctx, "test-overlay")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}

	if retrieved.ID != overlay.ID {
		t.Errorf("expected ID %q, got %q", overlay.ID, retrieved.ID)
	}

	// GetByName non-existent should fail
	_, err = store.GetByName(ctx, "non-existent")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDynamoDBStore_Update(t *testing.T) {
	client := newMockClient()
	store := NewDynamoDBStoreWithClient(client, "test-table")
	ctx := context.Background()

	overlay := entities.NewOverlay("test-overlay")
	_ = store.Create(ctx, overlay)

	// Add entities and update
	overlay.Universe.PutPrincipal(entities.Principal{Arn: "arn:aws:iam::123456789012:role/test"})
	err := store.Update(ctx, overlay)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify update
	retrieved, _ := store.Get(ctx, overlay.ID)
	if retrieved.NumPrincipals() != 1 {
		t.Errorf("expected 1 principal after update, got %d", retrieved.NumPrincipals())
	}

	// Update non-existent should fail
	nonExistent := entities.NewOverlay("non-existent")
	err = store.Update(ctx, nonExistent)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDynamoDBStore_Delete(t *testing.T) {
	client := newMockClient()
	store := NewDynamoDBStoreWithClient(client, "test-table")
	ctx := context.Background()

	overlay := entities.NewOverlay("test-overlay")
	_ = store.Create(ctx, overlay)

	// Delete should succeed
	err := store.Delete(ctx, overlay.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion
	_, err = store.Get(ctx, overlay.ID)
	if err != ErrNotFound {
		t.Error("overlay should be deleted")
	}

	// Delete non-existent should fail
	err = store.Delete(ctx, "non-existent")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDynamoDBStore_List(t *testing.T) {
	client := newMockClient()
	store := NewDynamoDBStoreWithClient(client, "test-table")
	ctx := context.Background()

	// Create multiple overlays
	overlay1 := entities.NewOverlay("alpha-overlay")
	overlay1.Universe.PutPrincipal(entities.Principal{Arn: "arn:aws:iam::123456789012:role/test"})
	_ = store.Create(ctx, overlay1)

	overlay2 := entities.NewOverlay("beta-overlay")
	overlay2.Universe.PutResource(entities.Resource{Arn: "arn:aws:s3:::test-bucket"})
	_ = store.Create(ctx, overlay2)

	overlay3 := entities.NewOverlay("gamma-test")
	_ = store.Create(ctx, overlay3)

	// List all
	summaries, err := store.List(ctx, "")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(summaries) != 3 {
		t.Errorf("expected 3 summaries, got %d", len(summaries))
	}

	// List with query
	summaries, err = store.List(ctx, "overlay")
	if err != nil {
		t.Fatalf("List with query failed: %v", err)
	}
	if len(summaries) != 2 {
		t.Errorf("expected 2 summaries matching 'overlay', got %d", len(summaries))
	}

	// Case-insensitive search
	summaries, err = store.List(ctx, "ALPHA")
	if err != nil {
		t.Fatalf("List with uppercase query failed: %v", err)
	}
	if len(summaries) != 1 {
		t.Errorf("expected 1 summary matching 'ALPHA', got %d", len(summaries))
	}
}

func TestDynamoDBStore_Exists(t *testing.T) {
	client := newMockClient()
	store := NewDynamoDBStoreWithClient(client, "test-table")
	ctx := context.Background()

	overlay := entities.NewOverlay("test-overlay")
	_ = store.Create(ctx, overlay)

	// Exists should return true
	exists, err := store.Exists(ctx, overlay.ID)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("expected overlay to exist")
	}

	// Non-existent should return false
	exists, err = store.Exists(ctx, "non-existent")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Error("expected overlay to not exist")
	}
}

func TestDynamoDBStore_DataIntegrity(t *testing.T) {
	client := newMockClient()
	store := NewDynamoDBStoreWithClient(client, "test-table")
	ctx := context.Background()

	// Create overlay with various entities
	overlay := entities.NewOverlay("complex-overlay")
	overlay.Universe.PutPrincipal(entities.Principal{
		Arn:  "arn:aws:iam::123456789012:role/test",
		Name: "test-role",
	})
	overlay.Universe.PutResource(entities.Resource{
		Arn:  "arn:aws:s3:::my-bucket",
		Type: "AWS::S3::Bucket",
	})
	overlay.Universe.PutPolicy(entities.ManagedPolicy{
		Arn:  "arn:aws:iam::123456789012:policy/test-policy",
		Name: "test-policy",
	})

	_ = store.Create(ctx, overlay)

	// Verify data integrity after round-trip
	retrieved, err := store.Get(ctx, overlay.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.NumPrincipals() != 1 {
		t.Errorf("expected 1 principal, got %d", retrieved.NumPrincipals())
	}
	if retrieved.NumResources() != 1 {
		t.Errorf("expected 1 resource, got %d", retrieved.NumResources())
	}
	if retrieved.NumPolicies() != 1 {
		t.Errorf("expected 1 policy, got %d", retrieved.NumPolicies())
	}

	// Verify summary counts
	summary := retrieved.Summary()
	if summary.NumPrincipals != 1 || summary.NumResources != 1 || summary.NumPolicies != 1 {
		t.Errorf("summary counts mismatch: P=%d R=%d Po=%d",
			summary.NumPrincipals, summary.NumResources, summary.NumPolicies)
	}
}

func TestItemToOverlay(t *testing.T) {
	store := NewDynamoDBStoreWithClient(nil, "test-table")

	overlay := entities.NewOverlay("test")
	overlay.Universe.PutPrincipal(entities.Principal{Arn: "arn:aws:iam::123456789012:role/test"})

	data := overlay.ToData()
	dataJSON, _ := json.Marshal(data)

	item := map[string]types.AttributeValue{
		attrID:        &types.AttributeValueMemberS{Value: overlay.ID},
		attrName:      &types.AttributeValueMemberS{Value: overlay.Name},
		attrCreatedAt: &types.AttributeValueMemberS{Value: overlay.CreatedAt.Format("2006-01-02T15:04:05.000Z")},
		attrData:      &types.AttributeValueMemberS{Value: string(dataJSON)},
	}

	result, err := store.itemToOverlay(item)
	if err != nil {
		t.Fatalf("itemToOverlay failed: %v", err)
	}

	if result.Name != overlay.Name {
		t.Errorf("expected name %q, got %q", overlay.Name, result.Name)
	}
	if result.NumPrincipals() != 1 {
		t.Errorf("expected 1 principal, got %d", result.NumPrincipals())
	}
}

func TestItemToOverlay_MissingData(t *testing.T) {
	store := NewDynamoDBStoreWithClient(nil, "test-table")

	// Item with no data attribute
	item := map[string]types.AttributeValue{
		attrID:   &types.AttributeValueMemberS{Value: "test-id"},
		attrName: &types.AttributeValueMemberS{Value: "test-name"},
	}

	_, err := store.itemToOverlay(item)
	if err == nil {
		t.Error("expected error for missing data attribute")
	}
}

func TestItemToOverlay_InvalidDataType(t *testing.T) {
	store := NewDynamoDBStoreWithClient(nil, "test-table")

	// Item with wrong type for data attribute
	item := map[string]types.AttributeValue{
		attrID:   &types.AttributeValueMemberS{Value: "test-id"},
		attrName: &types.AttributeValueMemberS{Value: "test-name"},
		attrData: &types.AttributeValueMemberN{Value: "123"}, // wrong type
	}

	_, err := store.itemToOverlay(item)
	if err == nil {
		t.Error("expected error for invalid data attribute type")
	}
}

func TestItemToOverlay_InvalidJSON(t *testing.T) {
	store := NewDynamoDBStoreWithClient(nil, "test-table")

	// Item with invalid JSON in data
	item := map[string]types.AttributeValue{
		attrID:   &types.AttributeValueMemberS{Value: "test-id"},
		attrName: &types.AttributeValueMemberS{Value: "test-name"},
		attrData: &types.AttributeValueMemberS{Value: "not valid json"},
	}

	_, err := store.itemToOverlay(item)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestNewStore_Memory(t *testing.T) {
	// Empty spec
	store, err := NewStore("")
	if err != nil {
		t.Fatalf("NewStore('') failed: %v", err)
	}
	if _, ok := store.(*MemoryStore); !ok {
		t.Error("expected MemoryStore for empty spec")
	}

	// Explicit memory
	store, err = NewStore("memory")
	if err != nil {
		t.Fatalf("NewStore('memory') failed: %v", err)
	}
	if _, ok := store.(*MemoryStore); !ok {
		t.Error("expected MemoryStore for 'memory' spec")
	}
}

func TestNewStore_InvalidSpec(t *testing.T) {
	_, err := NewStore("invalid://something")
	if err == nil {
		t.Error("expected error for invalid spec")
	}
}

func TestNewStore_DDBEmptyTable(t *testing.T) {
	_, err := NewStore("ddb://")
	if err == nil {
		t.Error("expected error for ddb:// with empty table name")
	}
}
