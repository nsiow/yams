package audit

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/nsiow/yams/cmd/yams/cli"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		wantErr string
	}{
		{
			name: "legacy actions",
			config: `[
				{"resource_type":"AWS::SQS::Queue","actions":["sqs:ReceiveMessage"]}
			]`,
		},
		{
			name: "action groups",
			config: `[
				{"resource_type":"AWS::SQS::Queue","action_groups":[
					{"name":"READ","actions":["sqs:ReceiveMessage"]}
				]}
			]`,
		},
		{
			name: "actions and groups",
			config: `[
				{"resource_type":"AWS::SQS::Queue","actions":["sqs:ReceiveMessage"],
				 "action_groups":[{"name":"READ","actions":["sqs:GetQueueAttributes"]}]}
			]`,
			wantErr: "must supply exactly one of actions or action_groups",
		},
		{
			name:    "no actions or groups",
			config:  `[{"resource_type":"AWS::SQS::Queue"}]`,
			wantErr: "must supply exactly one of actions or action_groups",
		},
		{
			name: "duplicate group",
			config: `[
				{"resource_type":"AWS::SQS::Queue","action_groups":[
					{"name":"READ","actions":["sqs:ReceiveMessage"]},
					{"name":"READ","actions":["sqs:GetQueueAttributes"]}
				]}
			]`,
			wantErr: "duplicate action group",
		},
		{
			name: "duplicate action canonical alias",
			config: `[
				{"resource_type":"AWS::SQS::Queue","action_groups":[
					{"name":"READ","actions":["sqs:ReceiveMessage"]},
					{"name":"WRITE","actions":["SQS.receivemessage"]}
				]}
			]`,
			wantErr: "belongs to both",
		},
		{
			name: "duplicate action in group",
			config: `[
				{"resource_type":"AWS::SQS::Queue","action_groups":[
					{"name":"READ","actions":["sqs:ReceiveMessage","SQS.receivemessage"]}
				]}
			]`,
			wantErr: "duplicate action",
		},
		{
			name: "unknown legacy action",
			config: `[
				{"resource_type":"AWS::SQS::Queue","actions":["sqs:NotAnAction"]}
			]`,
			wantErr: "unknown action",
		},
		{
			name: "unknown grouped action",
			config: `[
				{"resource_type":"AWS::SQS::Queue","action_groups":[
					{"name":"READ","actions":["sqs:NotAnAction"]}
				]}
			]`,
			wantErr: "unknown action",
		},
		{
			name: "missing group name",
			config: `[
				{"resource_type":"AWS::SQS::Queue","action_groups":[
					{"actions":["sqs:ReceiveMessage"]}
				]}
			]`,
			wantErr: "is missing name",
		},
		{
			name: "missing group actions",
			config: `[
				{"resource_type":"AWS::SQS::Queue","action_groups":[
					{"name":"READ"}
				]}
			]`,
			wantErr: "is missing actions",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.json")
			if err := os.WriteFile(path, []byte(tc.config), 0o600); err != nil {
				t.Fatal(err)
			}

			_, err := loadConfig(path)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("wanted error containing %q, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestLoadConfigCanonicalizesActions(t *testing.T) {
	config := `[
		{"resource_type":"AWS::SQS::Queue","actions":["SQS.receivemessage"]},
		{"resource_type":"AWS::SNS::Topic","action_groups":[
			{"name":"WRITE","actions":["SNS-PUBLISH"]}
		]}
	]`
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(config), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := loadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	want := []ConfigEntry{
		{
			ResourceType: "AWS::SQS::Queue",
			Actions:      []string{"sqs:ReceiveMessage"},
		},
		{
			ResourceType: "AWS::SNS::Topic",
			ActionGroups: []ActionGroup{{Name: "WRITE", Actions: []string{"sns:Publish"}}},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected config:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestValidateOutputOptions(t *testing.T) {
	legacy := []ConfigEntry{{ResourceType: "AWS::SQS::Queue", Actions: []string{"sqs:ReceiveMessage"}}}
	grouped := []ConfigEntry{{
		ResourceType: "AWS::SQS::Queue",
		ActionGroups: []ActionGroup{{Name: "READ", Actions: []string{"sqs:ReceiveMessage"}}},
	}}

	tests := []struct {
		name    string
		opts    cli.Flags
		config  []ConfigEntry
		wantErr bool
	}{
		{"legacy csv", cli.Flags{Format: formatCSV}, legacy, false},
		{"groups in legacy csv", cli.Flags{Format: formatCSV}, grouped, false},
		{"grouped csv", cli.Flags{Format: formatGroupedCSV, ResourceBatchSize: 1}, grouped, false},
		{"grouped jsonl", cli.Flags{Format: formatGroupedJSONL, ResourceBatchSize: 1}, grouped, false},
		{
			"legacy config grouped output",
			cli.Flags{Format: formatGroupedCSV, ResourceBatchSize: 1},
			legacy,
			true,
		},
		{"invalid batch size", cli.Flags{Format: formatGroupedCSV}, grouped, true},
		{"invalid format", cli.Flags{Format: "xml", ResourceBatchSize: 1}, grouped, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateOutputOptions(&tc.opts, tc.config)
			if (err != nil) != tc.wantErr {
				t.Fatalf("validateOutputOptions() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestCollapseS3Arn(t *testing.T) {
	if got := collapseS3Arn("arn:aws:s3:::example/key"); got != "arn:aws:s3:::example" {
		t.Fatalf("unexpected collapsed ARN: %s", got)
	}
	want := "arn:aws:sqs:us-east-1:111111111111:example"
	if got := collapseS3Arn(want); got != want {
		t.Fatalf("unexpected non-S3 ARN: %s", got)
	}
}
