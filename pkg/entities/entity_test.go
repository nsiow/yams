package entities

import (
	"testing"

	"github.com/nsiow/yams/pkg/policy"
)

// Test Account Key() and Repr()
func TestAccount_Key(t *testing.T) {
	a := Account{Id: "123456789012"}
	if a.Key() != "123456789012" {
		t.Fatalf("expected '123456789012' got '%s'", a.Key())
	}
}

func TestAccount_Repr(t *testing.T) {
	uv := NewUniverse()
	a := Account{Id: "123456789012"}
	uv.PutAccount(a)

	retrieved, _ := uv.Account(a.Id)
	repr, err := retrieved.Repr()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repr == nil {
		t.Fatal("expected non-nil repr")
	}
}

// Test Group Key() and Repr()
func TestGroup_Key(t *testing.T) {
	g := Group{Arn: "arn:aws:iam::123456789012:group/admin"}
	if g.Key() != "arn:aws:iam::123456789012:group/admin" {
		t.Fatalf("expected 'arn:aws:iam::123456789012:group/admin' got '%s'", g.Key())
	}
}

func TestGroup_Repr(t *testing.T) {
	uv := NewUniverse()
	g := Group{Arn: "arn:aws:iam::123456789012:group/admin"}
	uv.PutGroup(g)

	retrieved, _ := uv.Group(g.Arn)
	repr, err := retrieved.Repr()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repr == nil {
		t.Fatal("expected non-nil repr")
	}
}

// Test ManagedPolicy Key() and Repr()
func TestManagedPolicy_Key(t *testing.T) {
	p := ManagedPolicy{Arn: "arn:aws:iam::123456789012:policy/mypolicy"}
	if p.Key() != "arn:aws:iam::123456789012:policy/mypolicy" {
		t.Fatalf("expected 'arn:aws:iam::123456789012:policy/mypolicy' got '%s'", p.Key())
	}
}

func TestManagedPolicy_Repr(t *testing.T) {
	p := ManagedPolicy{
		Arn:    "arn:aws:iam::123456789012:policy/mypolicy",
		Policy: policy.Policy{},
	}
	repr, err := p.Repr()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repr == nil {
		t.Fatal("expected non-nil repr")
	}
}

// Test Principal Key() and Repr()
func TestPrincipal_Key(t *testing.T) {
	p := Principal{Arn: "arn:aws:iam::123456789012:role/myrole"}
	if p.Key() != "arn:aws:iam::123456789012:role/myrole" {
		t.Fatalf("expected 'arn:aws:iam::123456789012:role/myrole' got '%s'", p.Key())
	}
}

func TestPrincipal_Repr(t *testing.T) {
	uv := NewUniverse()
	p := Principal{Arn: "arn:aws:iam::123456789012:role/myrole"}
	uv.PutPrincipal(p)

	retrieved, _ := uv.Principal(p.Arn)
	repr, err := retrieved.Repr()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repr == nil {
		t.Fatal("expected non-nil repr")
	}
}

// Test Resource Key() and Repr()
func TestResource_Key(t *testing.T) {
	r := Resource{Arn: "arn:aws:s3:::mybucket"}
	if r.Key() != "arn:aws:s3:::mybucket" {
		t.Fatalf("expected 'arn:aws:s3:::mybucket' got '%s'", r.Key())
	}
}

func TestResource_Repr(t *testing.T) {
	uv := NewUniverse()
	r := Resource{Arn: "arn:aws:s3:::mybucket"}
	uv.PutResource(r)

	retrieved, _ := uv.Resource(r.Arn)
	repr, err := retrieved.Repr()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repr == nil {
		t.Fatal("expected non-nil repr")
	}
}

// Test Resource.Service()
func TestResource_Service(t *testing.T) {
	tests := []struct {
		name      string
		resType   string
		wantSvc   string
		shouldErr bool
	}{
		{"s3_bucket", "AWS::S3::Bucket", "s3", false},
		{"ec2_instance", "AWS::EC2::Instance", "ec2", false},
		{"iam_role", "AWS::IAM::Role", "iam", false},
		{"lambda_function", "AWS::Lambda::Function", "lambda", false},
		{"invalid_format", "InvalidType", "", true},
		{"too_few_parts", "AWS::S3", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := Resource{Type: tc.resType}
			svc, err := r.Service()

			if tc.shouldErr {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if svc != tc.wantSvc {
				t.Fatalf("expected '%s' got '%s'", tc.wantSvc, svc)
			}
		})
	}
}

// Test Resource.SubResource()
func TestResource_SubResource(t *testing.T) {
	uv := NewUniverse()
	bucket := Resource{
		uv:        uv,
		Type:      "AWS::S3::Bucket",
		Arn:       "arn:aws:s3:::mybucket",
		AccountId: "123456789012",
		Region:    "us-east-1",
		Tags:      []Tag{{Key: "env", Value: "prod"}},
		Policy:    policy.Policy{},
	}

	// Test S3 bucket subresource
	sub, err := bucket.SubResource("mykey/file.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sub.Arn != "arn:aws:s3:::mybucket/mykey/file.txt" {
		t.Fatalf("unexpected arn: %s", sub.Arn)
	}
	if sub.Type != "AWS::S3::Bucket::Object" {
		t.Fatalf("unexpected type: %s", sub.Type)
	}
	if sub.AccountId != bucket.AccountId {
		t.Fatalf("unexpected account: %s", sub.AccountId)
	}
	if sub.Tags != nil {
		t.Fatal("tags should not propagate to subresource")
	}

	// Test unsupported resource type
	ec2 := Resource{Type: "AWS::EC2::Instance", Arn: "arn:aws:ec2:us-east-1:123456789012:instance/i-123"}
	_, err = ec2.SubResource("something")
	if err == nil {
		t.Fatal("expected error for unsupported resource type")
	}
}

// Test subresource path handling
func TestResource_SubResource_PathNormalization(t *testing.T) {
	bucket := Resource{
		uv:   NewUniverse(),
		Type: "AWS::S3::Bucket",
		Arn:  "arn:aws:s3:::mybucket/",
	}

	sub, err := bucket.SubResource("/path/to/file")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "arn:aws:s3:::mybucket/path/to/file"
	if sub.Arn != expected {
		t.Fatalf("expected '%s' got '%s'", expected, sub.Arn)
	}
}
