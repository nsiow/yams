package awsconfig

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/nsiow/yams/pkg/entities"
)

// Loader provides the ability to load entity definitions from AWS Config data
type Loader struct {
	universe *entities.Universe
}

// NewLoader provisions and returns a new `Loader` struct, ready to use
func NewLoader() *Loader {
	return &Loader{
		universe: entities.NewUniverse(),
	}
}

// Universe returns an Universe containing the loaded Principals + Resources
func (l *Loader) Universe() *entities.Universe {
	return l.universe
}

// LoadJson loads data from a provided JSON array
// TODO(nsiow) consider having this load from io.Reader instead
func (l *Loader) LoadJson(data []byte) error {
	var items []ConfigItem
	err := json.Unmarshal(data, &items)
	if err != nil {
		return fmt.Errorf("unable to load data as JSON: %w", err)
	}
	return l.loadItems(items)
}

// LoadJson loads data from the provided newline-separate JSONL input
// TODO(nsiow) consider having this load from io.Reader instead
func (l *Loader) LoadJsonl(data []byte) error {
	r := bytes.NewReader(data)
	s := bufio.NewScanner(r)

	// Some buffer customization, since these JSON blobs can get big
	// TODO(nsiow) move these to constants
	buf := make([]byte, 0, 64*1024)
	s.Buffer(buf, 1024*1024)

	var items []ConfigItem
	for s.Scan() {
		// Read the next line; skip empty lines
		b := s.Bytes()
		if len(b) == 0 {
			continue
		}

		// Unmarshal into a single item
		var i ConfigItem
		err := json.Unmarshal(b, &i)
		if err != nil {
			return err
		}

		// Add to running list of items
		items = append(items, i)
	}

	// If we encountered an error scanning, return it
	if err := s.Err(); err != nil {
		return err
	}

	// Proceed to load
	return l.loadItems(items)
}

func (l *Loader) loadItems(items []ConfigItem) error {
	for _, item := range items {
		err := l.loadItem(item)
		if err != nil {
			return err
		}
	}

	return nil
}

func (l *Loader) loadItem(item ConfigItem) error {
	switch item.Type {
	case CONST_TYPE_YAMS_ORGANIZATIONS_ACCOUNT:
		return l.loadAccount(item)
	case CONST_TYPE_YAMS_ORGANIZATIONS_SCP:
		return l.loadSCP(item)
	case CONST_TYPE_AWS_IAM_GROUP:
		return l.loadGroup(item)
	case CONST_TYPE_AWS_IAM_POLICY:
		return l.loadManagedPolicy(item)
	case CONST_TYPE_AWS_IAM_ROLE:
		return l.loadRole(item)
	case CONST_TYPE_AWS_IAM_USER:
		return l.loadUser(item)
	case CONST_TYPE_AWS_S3_BUCKET:
		return l.loadBucket(item)
	case CONST_TYPE_AWS_DYNAMODB_TABLE:
		return l.loadTable(item)
	case CONST_TYPE_AWS_SNS_TOPIC:
		return l.loadTopic(item)
	case CONST_TYPE_AWS_SQS_QUEUE:
		return l.loadQueue(item)
	case CONST_TYPE_AWS_KMS_KEY:
		return l.loadKey(item)
	default:
		return l.loadGenericResource(item)
	}
}

func (l *Loader) loadAccount(item ConfigItem) error {
	var target configAccount

	err := json.Unmarshal(item.raw, &target)
	if err != nil {
		return err
	}

	l.universe.PutAccount(target.asAccount())
	return nil
}

func (l *Loader) loadSCP(item ConfigItem) error {
	var target configSCP

	err := json.Unmarshal(item.raw, &target)
	if err != nil {
		return err
	}

	l.universe.PutPolicy(target.asPolicy())
	return nil
}

func (l *Loader) loadGroup(item ConfigItem) error {
	var target configGroup

	err := json.Unmarshal(item.raw, &target)
	if err != nil {
		return err
	}

	l.universe.PutGroup(target.asGroup())
	return nil
}

func (l *Loader) loadManagedPolicy(item ConfigItem) error {
	var target configIamManagedPolicy

	err := json.Unmarshal(item.raw, &target)
	if err != nil {
		return err
	}

	resolvedPolicy, err := target.asPolicy()
	if err != nil {
		return err
	}

	l.universe.PutPolicy(resolvedPolicy)
	return nil
}

func (l *Loader) loadRole(item ConfigItem) error {
	var target configIamRole

	err := json.Unmarshal(item.raw, &target)
	if err != nil {
		return err
	}

	l.universe.PutResource(target.asResource())
	l.universe.PutPrincipal(target.asPrincipal())
	return nil
}

func (l *Loader) loadUser(item ConfigItem) error {
	var target configIamUser

	err := json.Unmarshal(item.raw, &target)
	if err != nil {
		return err
	}

	l.universe.PutResource(target.asResource())
	l.universe.PutPrincipal(target.asPrincipal())
	return nil
}

func (l *Loader) loadBucket(item ConfigItem) error {
	var target configS3Bucket

	err := json.Unmarshal(item.raw, &target)
	if err != nil {
		return err
	}

	l.universe.PutResource(target.asResource())
	return nil
}

func (l *Loader) loadTable(item ConfigItem) error {
	var target configDynamodbTable

	err := json.Unmarshal(item.raw, &target)
	if err != nil {
		return err
	}

	l.universe.PutResource(target.asResource())
	return nil
}

func (l *Loader) loadTopic(item ConfigItem) error {
	var target configSnsTopic

	err := json.Unmarshal(item.raw, &target)
	if err != nil {
		return err
	}

	l.universe.PutResource(target.asResource())
	return nil
}

func (l *Loader) loadQueue(item ConfigItem) error {
	var target configSqsQueue

	err := json.Unmarshal(item.raw, &target)
	if err != nil {
		return err
	}

	l.universe.PutResource(target.asResource())
	return nil
}

func (l *Loader) loadKey(item ConfigItem) error {
	var target configKmsKey

	err := json.Unmarshal(item.raw, &target)
	if err != nil {
		return err
	}

	l.universe.PutResource(target.asResource())
	return nil
}

func (l *Loader) loadGenericResource(item ConfigItem) error {
	var target genericResource

	err := json.Unmarshal(item.raw, &target)
	if err != nil {
		return err
	}

	l.universe.PutResource(target.asResource())
	return nil
}
