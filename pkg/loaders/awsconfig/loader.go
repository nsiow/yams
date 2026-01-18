package awsconfig

import (
	"bufio"
	"fmt"
	"io"

	json "github.com/bytedance/sonic"
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

// LoadJsonl loads data from the provided newline-separated JSONL input
func (l *Loader) LoadJsonl(reader io.Reader) error {
	s := bufio.NewScanner(reader)

	// Buffer customization for large JSON blobs
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
	var loadErr error
	l.uv.WithBulkWriter(func(w *entities.BulkWriter) {
		for _, blob := range blobs {
			err := l.loadItem(blob, w)
			if err != nil {
				loadErr = fmt.Errorf("error loading blob: %w\n%s", err, blob.raw)
				return
			}
		}
	})
	return loadErr
}

// -------------------------------------------------------------------------------------------------
// Load routing
// -------------------------------------------------------------------------------------------------

func (l *Loader) loadItem(blob configBlob, w *entities.BulkWriter) error {
	var err error

	switch blob.Type {
	case CONST_TYPE_YAMS_ORGANIZATIONS_ACCOUNT:
		err = l.loadAccount(blob, w)
	case CONST_TYPE_YAMS_ORGANIZATIONS_SCP:
		err = l.loadSCP(blob, w)
	case CONST_TYPE_YAMS_ORGANIZATIONS_RCP:
		err = l.loadRCP(blob, w)
	case CONST_TYPE_AWS_IAM_GROUP:
		err = l.loadGroup(blob, w)
	case CONST_TYPE_AWS_IAM_POLICY:
		err = l.loadManagedPolicy(blob, w)
	case CONST_TYPE_AWS_IAM_ROLE:
		err = l.loadRole(blob, w)
	case CONST_TYPE_AWS_IAM_USER:
		err = l.loadUser(blob, w)
	case CONST_TYPE_AWS_S3_BUCKET:
		err = l.loadBucket(blob, w)
	case CONST_TYPE_AWS_DYNAMODB_TABLE:
		err = l.loadTable(blob, w)
	case CONST_TYPE_AWS_SNS_TOPIC:
		err = l.loadTopic(blob, w)
	case CONST_TYPE_AWS_SQS_QUEUE:
		err = l.loadQueue(blob, w)
	case CONST_TYPE_AWS_KMS_KEY:
		err = l.loadKey(blob, w)
	default:
		err = l.loadGenericResource(blob, w)
	}

	if err != nil {
		return fmt.Errorf("error while loading item of type '%s': %w", blob.Type, err)
	}

	return nil
}

func (l *Loader) loadAccount(blob configBlob, w *entities.BulkWriter) error {
	var target Account

	err := json.Unmarshal(blob.raw, &target)
	if err != nil {
		return err
	}

	w.PutAccount(target.asAccount())
	return nil
}

func (l *Loader) loadSCP(blob configBlob, w *entities.BulkWriter) error {
	var target SCP

	err := json.Unmarshal(blob.raw, &target)
	if err != nil {
		return err
	}

	w.PutPolicy(target.asPolicy())
	w.PutResource(target.asResource())
	return nil
}

func (l *Loader) loadRCP(blob configBlob, w *entities.BulkWriter) error {
	var target RCP

	err := json.Unmarshal(blob.raw, &target)
	if err != nil {
		return err
	}

	w.PutPolicy(target.asPolicy())
	w.PutResource(target.asResource())
	return nil
}

func (l *Loader) loadGroup(blob configBlob, w *entities.BulkWriter) error {
	var target IamGroup

	err := json.Unmarshal(blob.raw, &target)
	if err != nil {
		return err
	}

	w.PutGroup(target.asGroup())
	w.PutResource(target.asResource())
	return nil
}

func (l *Loader) loadManagedPolicy(blob configBlob, w *entities.BulkWriter) error {
	var target IamPolicy

	err := json.Unmarshal(blob.raw, &target)
	if err != nil {
		return err
	}

	resolvedPolicy, err := target.asPolicy()
	if err != nil {
		return err
	}

	w.PutPolicy(resolvedPolicy)
	w.PutResource(target.asResource())
	return nil
}

func (l *Loader) loadRole(blob configBlob, w *entities.BulkWriter) error {
	var target IamRole

	err := json.Unmarshal(blob.raw, &target)
	if err != nil {
		return err
	}

	w.PutPrincipal(target.asPrincipal())
	w.PutResource(target.asResource())
	return nil
}

func (l *Loader) loadUser(blob configBlob, w *entities.BulkWriter) error {
	var target IamUser

	err := json.Unmarshal(blob.raw, &target)
	if err != nil {
		return err
	}

	w.PutPrincipal(target.asPrincipal())
	w.PutResource(target.asResource())
	return nil
}

func (l *Loader) loadBucket(blob configBlob, w *entities.BulkWriter) error {
	var target S3Bucket

	err := json.Unmarshal(blob.raw, &target)
	if err != nil {
		return err
	}

	w.PutResource(target.asResource())
	return nil
}

func (l *Loader) loadTable(blob configBlob, w *entities.BulkWriter) error {
	var target DynamodbTable

	err := json.Unmarshal(blob.raw, &target)
	if err != nil {
		return err
	}

	w.PutResource(target.asResource())
	return nil
}

func (l *Loader) loadTopic(blob configBlob, w *entities.BulkWriter) error {
	var target SnsTopic

	err := json.Unmarshal(blob.raw, &target)
	if err != nil {
		return err
	}

	w.PutResource(target.asResource())
	return nil
}

func (l *Loader) loadQueue(blob configBlob, w *entities.BulkWriter) error {
	var target SqsQueue

	err := json.Unmarshal(blob.raw, &target)
	if err != nil {
		return err
	}

	w.PutResource(target.asResource())
	return nil
}

func (l *Loader) loadKey(blob configBlob, w *entities.BulkWriter) error {
	var target KmsKey

	err := json.Unmarshal(blob.raw, &target)
	if err != nil {
		return err
	}

	w.PutResource(target.asResource())
	return nil
}

func (l *Loader) loadGenericResource(blob configBlob, w *entities.BulkWriter) error {
	var target genericResource

	err := json.Unmarshal(blob.raw, &target)
	if err != nil {
		return err
	}

	w.PutResource(target.asResource())
	return nil
}
