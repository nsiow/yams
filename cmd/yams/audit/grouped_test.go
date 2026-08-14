package audit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
	"github.com/nsiow/yams/pkg/sim"
)

type capturedGroup struct {
	resource   string
	group      string
	principals []string
}

type capturingGroupedWriter struct {
	groups []capturedGroup
}

func (w *capturingGroupedWriter) WriteGroup(resource, group string, principals []string) error {
	w.groups = append(w.groups, capturedGroup{
		resource:   resource,
		group:      group,
		principals: principals,
	})
	return nil
}

func (w *capturingGroupedWriter) Flush() error {
	return nil
}

func TestProcessGroupedEntry(t *testing.T) {
	const accountID = "111111111111"
	readPrincipalArn := "arn:aws:iam::" + accountID + ":role/a-read"
	writePrincipalArn := "arn:aws:iam::" + accountID + ":role/b-write"
	queueA := "arn:aws:sqs:us-east-1:" + accountID + ":a-queue"
	queueB := "arn:aws:sqs:us-east-1:" + accountID + ":b-queue"

	readPrincipal := entities.Principal{
		Type:      "AWS::IAM::Role",
		AccountId: accountID,
		Arn:       readPrincipalArn,
		InlinePolicies: []policy.Policy{{
			Statement: []policy.Statement{{
				Effect:   policy.EFFECT_ALLOW,
				Action:   []string{"sqs:ReceiveMessage", "sqs:GetQueueAttributes"},
				Resource: []string{queueA},
			}},
		}},
	}
	writePrincipal := entities.Principal{
		Type:      "AWS::IAM::Role",
		AccountId: accountID,
		Arn:       writePrincipalArn,
		InlinePolicies: []policy.Policy{{
			Statement: []policy.Statement{{
				Effect:   policy.EFFECT_ALLOW,
				Action:   []string{"sqs:SendMessage"},
				Resource: []string{queueA},
			}},
		}},
	}

	simulator, err := sim.NewSimulator()
	if err != nil {
		t.Fatal(err)
	}
	simulator.Universe = entities.NewBuilder().
		WithPrincipals(writePrincipal, readPrincipal).
		WithResources(
			entities.Resource{Type: "AWS::SQS::Queue", AccountId: accountID, Arn: queueB},
			entities.Resource{Type: "AWS::SQS::Queue", AccountId: accountID, Arn: queueA},
		).
		Build()

	opts := sim.NewOptions()
	frozenPrincipals, err := simulator.FreezePrincipals(
		[]string{readPrincipalArn, writePrincipalArn},
		opts,
	)
	if err != nil {
		t.Fatal(err)
	}

	entry := ConfigEntry{
		ResourceType: "AWS::SQS::Queue",
		ActionGroups: []ActionGroup{
			{Name: "READ", Actions: []string{"sqs:ReceiveMessage", "sqs:GetQueueAttributes"}},
			{Name: "WRITE", Actions: []string{"sqs:SendMessage"}},
		},
	}
	want := []capturedGroup{
		{resource: queueA, group: "READ", principals: []string{readPrincipalArn}},
		{resource: queueA, group: "WRITE", principals: []string{writePrincipalArn}},
		{resource: queueB, group: "READ", principals: []string{}},
		{resource: queueB, group: "WRITE", principals: []string{}},
	}
	for _, batchSize := range []int{1, 256} {
		t.Run(fmt.Sprintf("batch_size_%d", batchSize), func(t *testing.T) {
			writer := &capturingGroupedWriter{}
			groups, relationships, err := processGroupedEntry(
				simulator,
				frozenPrincipals,
				entry,
				opts,
				batchSize,
				writer,
			)
			if err != nil {
				t.Fatal(err)
			}
			if groups != 4 || relationships != 2 {
				t.Fatalf(
					"got groups=%d relationships=%d, want groups=4 relationships=2",
					groups,
					relationships,
				)
			}
			if !reflect.DeepEqual(writer.groups, want) {
				t.Fatalf("unexpected groups:\n got: %#v\nwant: %#v", writer.groups, want)
			}
		})
	}
}

func TestProcessGroupedEntryCollapsesS3Resources(t *testing.T) {
	const accountID = "111111111111"
	principalArn := "arn:aws:iam::" + accountID + ":role/reader"
	bucketArn := "arn:aws:s3:::example-bucket"

	principal := entities.Principal{
		Type:      "AWS::IAM::Role",
		AccountId: accountID,
		Arn:       principalArn,
		InlinePolicies: []policy.Policy{{
			Statement: []policy.Statement{{
				Effect:   policy.EFFECT_ALLOW,
				Action:   []string{"s3:GetObject", "s3:ListBucket"},
				Resource: []string{bucketArn, bucketArn + "/*"},
			}},
		}},
	}

	simulator, err := sim.NewSimulator()
	if err != nil {
		t.Fatal(err)
	}
	simulator.Universe = entities.NewBuilder().
		WithPrincipals(principal).
		WithResources(entities.Resource{
			Type:      "AWS::S3::Bucket",
			AccountId: accountID,
			Arn:       bucketArn,
		}).
		Build()

	opts := sim.NewOptions()
	frozenPrincipals, err := simulator.FreezePrincipals([]string{principalArn}, opts)
	if err != nil {
		t.Fatal(err)
	}
	entry := ConfigEntry{
		ResourceType: "AWS::S3::Bucket",
		ActionGroups: []ActionGroup{{
			Name:    "READ",
			Actions: []string{"s3:GetObject", "s3:ListBucket"},
		}},
	}
	writer := &capturingGroupedWriter{}

	groups, relationships, err := processGroupedEntry(
		simulator,
		frozenPrincipals,
		entry,
		opts,
		1,
		writer,
	)
	if err != nil {
		t.Fatal(err)
	}
	if groups != 1 || relationships != 1 {
		t.Fatalf("got groups=%d relationships=%d, want groups=1 relationships=1", groups, relationships)
	}
	want := []capturedGroup{{
		resource:   bucketArn,
		group:      "READ",
		principals: []string{principalArn},
	}}
	if !reflect.DeepEqual(writer.groups, want) {
		t.Fatalf("unexpected groups:\n got: %#v\nwant: %#v", writer.groups, want)
	}
}

func TestPrincipalsFromBitsAcrossWords(t *testing.T) {
	frozenPrincipals := make([]*entities.FrozenPrincipal, 66)
	for i := range frozenPrincipals {
		frozenPrincipals[i] = &entities.FrozenPrincipal{
			Arn: fmt.Sprintf("arn:aws:iam::111111111111:role/principal-%02d", i),
		}
	}

	got := principalsFromBits([]uint64{1 << 63, 0b11}, frozenPrincipals)
	want := []string{
		frozenPrincipals[63].Arn,
		frozenPrincipals[64].Arn,
		frozenPrincipals[65].Arn,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected principals:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestGroupedCSVWriter(t *testing.T) {
	var output bytes.Buffer
	w, err := newGroupedCSVWriter(&output)
	if err != nil {
		t.Fatal(err)
	}
	if err := w.WriteGroup("resource-a", "READ", []string{"principal-a", "principal-b"}); err != nil {
		t.Fatal(err)
	}
	if err := w.WriteGroup("resource-b", "WRITE", nil); err != nil {
		t.Fatal(err)
	}
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}

	want := "resource,group,principal\n" +
		"resource-a,READ,principal-a\n" +
		"resource-a,READ,principal-b\n" +
		"resource-b,WRITE,\n"
	if output.String() != want {
		t.Fatalf("unexpected CSV:\n%s", output.String())
	}
}

func TestGroupedJSONLWriter(t *testing.T) {
	var output bytes.Buffer
	w := &groupedJSONLWriter{w: json.NewEncoder(&output)}
	if err := w.WriteGroup("resource-a", "READ", []string{"principal-a"}); err != nil {
		t.Fatal(err)
	}
	if err := w.WriteGroup("resource-b", "WRITE", nil); err != nil {
		t.Fatal(err)
	}

	want := "{\"resource\":\"resource-a\",\"group\":\"READ\",\"principals\":[\"principal-a\"]}\n" +
		"{\"resource\":\"resource-b\",\"group\":\"WRITE\",\"principals\":[]}\n"
	if output.String() != want {
		t.Fatalf("unexpected JSONL:\n%s", output.String())
	}
}
