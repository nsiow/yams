package audit

import (
	"bytes"
	"encoding/json"
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
	if groups != 4 || relationships != 2 {
		t.Fatalf("got groups=%d relationships=%d, want groups=4 relationships=2", groups, relationships)
	}

	want := []capturedGroup{
		{resource: queueA, group: "READ", principals: []string{readPrincipalArn}},
		{resource: queueA, group: "WRITE", principals: []string{writePrincipalArn}},
		{resource: queueB, group: "READ", principals: []string{}},
		{resource: queueB, group: "WRITE", principals: []string{}},
	}
	if !reflect.DeepEqual(writer.groups, want) {
		t.Fatalf("unexpected groups:\n got: %#v\nwant: %#v", writer.groups, want)
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
