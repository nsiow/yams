package awsconfig

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"

	"github.com/nsiow/yams/pkg/entities"
)

const SCAN_BUF_SIZE = 1024 * 1024

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

// LoadJsonl loads data from the provided newline-separate JSONL input
func (l *Loader) LoadJsonl(reader io.Reader) error {
	s := bufio.NewScanner(reader)

	// Some buffer customization, since these JSON blobs can get big
	// TODO(nsiow) move these to constants
	buf := make([]byte, SCAN_BUF_SIZE)
	s.Buffer(buf, len(buf))

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
			return fmt.Errorf("error decoding fragment: %w", err)
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
			return fmt.Errorf("error loading blob: %w\n%s", err, blob.raw)
		}
	}

	return nil
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
