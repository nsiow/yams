package awsconfig

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"

	"github.com/nsiow/yams/pkg/entities"
)

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

// LoadJson loads data from a provided JSON array
func (l *Loader) LoadJson(reader io.Reader) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	var blobs []configBlob
	err = json.Unmarshal(data, &blobs)
	if err != nil {
		return fmt.Errorf("unable to load data as JSON: %w", err)
	}
	return l.loadItems(blobs)
}

// LoadJson loads data from the provided newline-separate JSONL input
func (l *Loader) LoadJsonl(reader io.Reader) error {
	s := bufio.NewScanner(reader)

	// Some buffer customization, since these JSON blobs can get big
	// TODO(nsiow) move these to constants
	buf := make([]byte, 0, 64*1024)
	s.Buffer(buf, 1024*1024)

	var blobs []configBlob
	for s.Scan() {
		// Read the next line; skip empty lines
		b := s.Bytes()
		if len(b) == 0 {
			continue
		}

		// Unmarshal into a single item
		var i configBlob
		err := json.Unmarshal(b, &i)
		if err != nil {
			return err
		}

		// Add to running list of items
		blobs = append(blobs, i)
	}

	// If we encountered an error scanning, return it
	if err := s.Err(); err != nil {
		return err
	}

	// Proceed to load
	return l.loadItems(blobs)
}

func (l *Loader) loadItems(blobs []configBlob) error {
	for _, blob := range blobs {
		err := l.loadItem(blob)
		if err != nil {
			return err
		}
	}

	return nil
}

// -------------------------------------------------------------------------------------------------
// Load routing
// -------------------------------------------------------------------------------------------------

func (l *Loader) loadItem(blob configBlob) error {
	switch blob.Type {
	case CONST_TYPE_YAMS_ORGANIZATIONS_ACCOUNT:
		return l.loadAccount(blob)
	case CONST_TYPE_YAMS_ORGANIZATIONS_SCP:
		return l.loadSCP(blob)
	case CONST_TYPE_AWS_IAM_GROUP:
		return l.loadGroup(blob)
	case CONST_TYPE_AWS_IAM_POLICY:
		return l.loadManagedPolicy(blob)
	case CONST_TYPE_AWS_IAM_ROLE:
		return l.loadRole(blob)
	case CONST_TYPE_AWS_IAM_USER:
		return l.loadUser(blob)
	case CONST_TYPE_AWS_S3_BUCKET:
		return l.loadBucket(blob)
	case CONST_TYPE_AWS_DYNAMODB_TABLE:
		return l.loadTable(blob)
	case CONST_TYPE_AWS_SNS_TOPIC:
		return l.loadTopic(blob)
	case CONST_TYPE_AWS_SQS_QUEUE:
		return l.loadQueue(blob)
	case CONST_TYPE_AWS_KMS_KEY:
		return l.loadKey(blob)
	default:
		return l.loadGenericResource(blob)
	}
}

func (l *Loader) loadAccount(blob configBlob) error {
	var target configAccount

	err := json.Unmarshal(blob.raw, &target)
	if err != nil {
		return err
	}

	l.uv.PutAccount(target.asAccount())
	return nil
}

func (l *Loader) loadSCP(blob configBlob) error {
	var target configSCP

	err := json.Unmarshal(blob.raw, &target)
	if err != nil {
		return err
	}

	l.uv.PutPolicy(target.asPolicy())
	l.uv.PutResource(target.asResource())
	return nil
}

func (l *Loader) loadGroup(blob configBlob) error {
	var target configGroup

	err := json.Unmarshal(blob.raw, &target)
	if err != nil {
		return err
	}

	l.uv.PutGroup(target.asGroup())
	l.uv.PutResource(target.asResource())
	return nil
}

func (l *Loader) loadManagedPolicy(blob configBlob) error {
	var target configIamManagedPolicy

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
	var target configIamRole

	err := json.Unmarshal(blob.raw, &target)
	if err != nil {
		return err
	}

	l.uv.PutPrincipal(target.asPrincipal())
	l.uv.PutResource(target.asResource())
	return nil
}

func (l *Loader) loadUser(blob configBlob) error {
	var target configIamUser

	err := json.Unmarshal(blob.raw, &target)
	if err != nil {
		return err
	}

	l.uv.PutPrincipal(target.asPrincipal())
	l.uv.PutResource(target.asResource())
	return nil
}

func (l *Loader) loadBucket(blob configBlob) error {
	var target configS3Bucket

	err := json.Unmarshal(blob.raw, &target)
	if err != nil {
		return err
	}

	l.uv.PutResource(target.asResource())
	return nil
}

func (l *Loader) loadTable(blob configBlob) error {
	var target configDynamodbTable

	err := json.Unmarshal(blob.raw, &target)
	if err != nil {
		return err
	}

	l.uv.PutResource(target.asResource())
	return nil
}

func (l *Loader) loadTopic(blob configBlob) error {
	var target configSnsTopic

	err := json.Unmarshal(blob.raw, &target)
	if err != nil {
		return err
	}

	l.uv.PutResource(target.asResource())
	return nil
}

func (l *Loader) loadQueue(blob configBlob) error {
	var target configSqsQueue

	err := json.Unmarshal(blob.raw, &target)
	if err != nil {
		return err
	}

	l.uv.PutResource(target.asResource())
	return nil
}

func (l *Loader) loadKey(blob configBlob) error {
	var target configKmsKey

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
