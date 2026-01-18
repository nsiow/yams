package awsconfig

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/nsiow/yams/pkg/entities"
)

const SCAN_BUF_SIZE = 1024 * 1024

// resourceTypeKey is the JSON key we search for when extracting the resource type
var resourceTypeKey = []byte(`"resourceType"`)

// extractResourceType quickly extracts the resourceType from raw JSON without full parsing.
// Returns empty string if not found.
func extractResourceType(data []byte) string {
	idx := bytes.Index(data, resourceTypeKey)
	if idx == -1 {
		return ""
	}

	// Skip past the key and find the colon
	start := idx + len(resourceTypeKey)
	for start < len(data) && (data[start] == ' ' || data[start] == '\t' || data[start] == ':') {
		start++
	}

	// Expect opening quote
	if start >= len(data) || data[start] != '"' {
		return ""
	}
	start++ // skip opening quote

	// Find closing quote
	end := start
	for end < len(data) && data[end] != '"' {
		end++
	}

	return string(data[start:end])
}

// Loader provides the ability to load entity definitions from AWS Config data
type Loader struct {
	uv *entities.Universe
}

// NewLoader provisions and returns a new `Loader` struct, ready to use
func NewLoader() *Loader {
	return &Loader{
		uv: entities.NewUniverse(),
	}
}

// Universe returns an Universe containing the loaded Principals + Resources
func (l *Loader) Universe() *entities.Universe {
	return l.uv
}

// LoadJson loads data from a provided JSON array using streaming decoding
func (l *Loader) LoadJson(reader io.Reader) error {
	dec := json.NewDecoder(reader)

	// Consume opening bracket
	tok, err := dec.Token()
	if err != nil {
		return fmt.Errorf("error reading opening token: %w", err)
	}
	if tok != json.Delim('[') {
		return fmt.Errorf("expected JSON array, got %v", tok)
	}

	// Stream and process items in parallel
	return l.loadJsonParallel(dec)
}

// loadJsonParallel streams JSON array elements and processes them in parallel
func (l *Loader) loadJsonParallel(dec *json.Decoder) error {
	pool := newLoaderPool(l)
	pool.start()

	// Stream each array element
	for dec.More() {
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			pool.close()
			return fmt.Errorf("error decoding array element: %w", err)
		}

		// Fast type extraction without full JSON parsing
		typ := extractResourceType(raw)
		pool.submit(typ, raw)
	}

	pool.close()
	return pool.error()
}

// LoadJsonl loads data from the provided newline-separated JSONL input
func (l *Loader) LoadJsonl(reader io.Reader) error {
	s := bufio.NewScanner(reader)

	// Buffer customization for large JSON blobs
	buf := make([]byte, SCAN_BUF_SIZE)
	s.Buffer(buf, len(buf))

	pool := newLoaderPool(l)
	pool.start()

	for s.Scan() {
		// Read the next line; skip empty lines
		b := s.Bytes()
		if len(b) == 0 {
			continue
		}

		// Fast type extraction without full JSON parsing
		typ := extractResourceType(b)

		// Copy the bytes since scanner reuses buffer
		raw := make(json.RawMessage, len(b))
		copy(raw, b)

		pool.submit(typ, raw)
	}

	// If we encountered an error scanning, return it
	if err := s.Err(); err != nil {
		pool.close()
		return err
	}

	pool.close()
	return pool.error()
}

// -------------------------------------------------------------------------------------------------
// Load routing
// -------------------------------------------------------------------------------------------------

func (l *Loader) loadItem(blob configBlob) error {
	var err error

	switch blob.Type {
	case CONST_TYPE_YAMS_ORGANIZATIONS_ACCOUNT:
		err = l.loadAccount(blob)
	case CONST_TYPE_YAMS_ORGANIZATIONS_SCP:
		err = l.loadSCP(blob)
	case CONST_TYPE_YAMS_ORGANIZATIONS_RCP:
		err = l.loadRCP(blob)
	case CONST_TYPE_AWS_IAM_GROUP:
		err = l.loadGroup(blob)
	case CONST_TYPE_AWS_IAM_POLICY:
		err = l.loadManagedPolicy(blob)
	case CONST_TYPE_AWS_IAM_ROLE:
		err = l.loadRole(blob)
	case CONST_TYPE_AWS_IAM_USER:
		err = l.loadUser(blob)
	case CONST_TYPE_AWS_S3_BUCKET:
		err = l.loadBucket(blob)
	case CONST_TYPE_AWS_DYNAMODB_TABLE:
		err = l.loadTable(blob)
	case CONST_TYPE_AWS_SNS_TOPIC:
		err = l.loadTopic(blob)
	case CONST_TYPE_AWS_SQS_QUEUE:
		err = l.loadQueue(blob)
	case CONST_TYPE_AWS_KMS_KEY:
		err = l.loadKey(blob)
	default:
		err = l.loadGenericResource(blob)
	}

	if err != nil {
		return fmt.Errorf("error while loading item of type '%s': %w", blob.Type, err)
	}

	return nil
}

func (l *Loader) loadAccount(blob configBlob) error {
	var target Account

	err := json.Unmarshal(blob.raw, &target)
	if err != nil {
		return err
	}

	l.uv.PutAccount(target.asAccount())
	return nil
}

func (l *Loader) loadSCP(blob configBlob) error {
	var target SCP

	err := json.Unmarshal(blob.raw, &target)
	if err != nil {
		return err
	}

	l.uv.PutPolicy(target.asPolicy())
	l.uv.PutResource(target.asResource())
	return nil
}

func (l *Loader) loadRCP(blob configBlob) error {
	var target RCP

	err := json.Unmarshal(blob.raw, &target)
	if err != nil {
		return err
	}

	l.uv.PutPolicy(target.asPolicy())
	l.uv.PutResource(target.asResource())
	return nil
}

func (l *Loader) loadGroup(blob configBlob) error {
	var target IamGroup

	err := json.Unmarshal(blob.raw, &target)
	if err != nil {
		return err
	}

	l.uv.PutGroup(target.asGroup())
	l.uv.PutResource(target.asResource())
	return nil
}

func (l *Loader) loadManagedPolicy(blob configBlob) error {
	var target IamPolicy

	err := json.Unmarshal(blob.raw, &target)
	if err != nil {
		return err
	}

	resolvedPolicy, err := target.asPolicy()
	if err != nil {
		return err
	}

	l.uv.PutPolicy(resolvedPolicy)
	l.uv.PutResource(target.asResource())
	return nil
}

func (l *Loader) loadRole(blob configBlob) error {
	var target IamRole

	err := json.Unmarshal(blob.raw, &target)
	if err != nil {
		return err
	}

	l.uv.PutPrincipal(target.asPrincipal())
	l.uv.PutResource(target.asResource())
	return nil
}

func (l *Loader) loadUser(blob configBlob) error {
	var target IamUser

	err := json.Unmarshal(blob.raw, &target)
	if err != nil {
		return err
	}

	l.uv.PutPrincipal(target.asPrincipal())
	l.uv.PutResource(target.asResource())
	return nil
}

func (l *Loader) loadBucket(blob configBlob) error {
	var target S3Bucket

	err := json.Unmarshal(blob.raw, &target)
	if err != nil {
		return err
	}

	l.uv.PutResource(target.asResource())
	return nil
}

func (l *Loader) loadTable(blob configBlob) error {
	var target DynamodbTable

	err := json.Unmarshal(blob.raw, &target)
	if err != nil {
		return err
	}

	l.uv.PutResource(target.asResource())
	return nil
}

func (l *Loader) loadTopic(blob configBlob) error {
	var target SnsTopic

	err := json.Unmarshal(blob.raw, &target)
	if err != nil {
		return err
	}

	l.uv.PutResource(target.asResource())
	return nil
}

func (l *Loader) loadQueue(blob configBlob) error {
	var target SqsQueue

	err := json.Unmarshal(blob.raw, &target)
	if err != nil {
		return err
	}

	l.uv.PutResource(target.asResource())
	return nil
}

func (l *Loader) loadKey(blob configBlob) error {
	var target KmsKey

	err := json.Unmarshal(blob.raw, &target)
	if err != nil {
		return err
	}

	l.uv.PutResource(target.asResource())
	return nil
}

func (l *Loader) loadGenericResource(blob configBlob) error {
	var target genericResource

	err := json.Unmarshal(blob.raw, &target)
	if err != nil {
		return err
	}

	l.uv.PutResource(target.asResource())
	return nil
}
