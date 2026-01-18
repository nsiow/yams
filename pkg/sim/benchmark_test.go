package sim

import (
	"compress/gzip"
	"os"
	"testing"

	"github.com/nsiow/yams/pkg/aws/sar"
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/loaders/awsconfig"
	"github.com/nsiow/yams/pkg/policy"
	"github.com/nsiow/yams/pkg/sim/wildcard"
)

// -------------------------------------------------------------------------------------------------
// Wildcard Matching Benchmarks
// -------------------------------------------------------------------------------------------------

func BenchmarkWildcard_ExactMatch(b *testing.B) {
	pattern := "arn:aws:s3:::my-bucket"
	value := "arn:aws:s3:::my-bucket"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wildcard.MatchSegments(pattern, value)
	}
}

func BenchmarkWildcard_PrefixMatch(b *testing.B) {
	pattern := "arn:aws:s3:::my-bucket/*"
	value := "arn:aws:s3:::my-bucket/object.txt"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wildcard.MatchSegments(pattern, value)
	}
}

func BenchmarkWildcard_SuffixMatch(b *testing.B) {
	pattern := "arn:aws:iam::*:role/MyRole"
	value := "arn:aws:iam::123456789012:role/MyRole"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wildcard.MatchSegments(pattern, value)
	}
}

func BenchmarkWildcard_MiddleMatch(b *testing.B) {
	pattern := "arn:aws:s3:::*bucket*"
	value := "arn:aws:s3:::my-bucket-name"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wildcard.MatchSegments(pattern, value)
	}
}

func BenchmarkWildcard_RegexFallback(b *testing.B) {
	pattern := "arn:aws:s3:::my-*-bucket-?"
	value := "arn:aws:s3:::my-test-bucket-1"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wildcard.MatchSegments(pattern, value)
	}
}

func BenchmarkWildcard_ComplexRegex(b *testing.B) {
	pattern := "arn:aws:iam::*:role/*-?-prod-*"
	value := "arn:aws:iam::123456789012:role/app-1-prod-service"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wildcard.MatchSegments(pattern, value)
	}
}

func BenchmarkWildcard_IgnoreCase(b *testing.B) {
	pattern := "s3:GetObject"
	value := "S3:GETOBJECT"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wildcard.MatchSegmentsIgnoreCase(pattern, value)
	}
}

func BenchmarkWildcard_ArnMatch(b *testing.B) {
	pattern := "arn:aws:iam::123456789012:role/*"
	value := "arn:aws:iam::123456789012:role/MyRole"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wildcard.MatchArn(pattern, value)
	}
}

// -------------------------------------------------------------------------------------------------
// Statement Evaluation Benchmarks
// -------------------------------------------------------------------------------------------------

func createTestSubject() *subject {
	action, _ := sar.LookupString("s3:GetObject")
	principal := &entities.FrozenPrincipal{
		Type:      "AWS::IAM::Role",
		AccountId: "123456789012",
		Arn:       "arn:aws:iam::123456789012:role/TestRole",
	}
	resource := &entities.FrozenResource{
		Type:      "AWS::S3::Bucket",
		AccountId: "123456789012",
		Arn:       "arn:aws:s3:::test-bucket/object.txt",
	}

	ac := AuthContext{
		Action:    action,
		Principal: principal,
		Resource:  resource,
	}

	return newSubject(ac, DEFAULT_OPTIONS)
}

func BenchmarkEvalStatement_ActionMatch(b *testing.B) {
	subj := createTestSubject()
	stmt := policy.Statement{
		Effect: policy.EFFECT_ALLOW,
		Action: policy.Action{"s3:*"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evalStatementMatchesAction(subj, &stmt)
	}
}

func BenchmarkEvalStatement_ActionMatchMultiple(b *testing.B) {
	subj := createTestSubject()
	stmt := policy.Statement{
		Effect: policy.EFFECT_ALLOW,
		Action: policy.Action{"s3:PutObject", "s3:DeleteObject", "s3:GetObject", "s3:ListBucket"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evalStatementMatchesAction(subj, &stmt)
	}
}

func BenchmarkEvalStatement_ResourceMatch(b *testing.B) {
	subj := createTestSubject()
	stmt := policy.Statement{
		Effect:   policy.EFFECT_ALLOW,
		Resource: policy.Resource{"arn:aws:s3:::test-bucket/*"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evalStatementMatchesResource(subj, &stmt)
	}
}

func BenchmarkEvalStatement_PrincipalMatch(b *testing.B) {
	subj := createTestSubject()
	stmt := policy.Statement{
		Effect: policy.EFFECT_ALLOW,
		Principal: policy.Principal{
			AWS: policy.Value{"arn:aws:iam::123456789012:role/TestRole"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evalStatementMatchesPrincipal(subj, &stmt)
	}
}

func BenchmarkEvalStatement_ConditionStringEquals(b *testing.B) {
	subj := createTestSubject()
	subj.auth.Properties = NewBagFromMap(map[string]string{
		"aws:RequestedRegion": "us-east-1",
	})

	stmt := policy.Statement{
		Effect: policy.EFFECT_ALLOW,
		Condition: map[string]policy.ConditionValues{
			"StringEquals": {
				"aws:RequestedRegion": {"us-east-1"},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evalStatementMatchesCondition(subj, &stmt)
	}
}

func BenchmarkEvalStatement_ConditionIpAddress(b *testing.B) {
	subj := createTestSubject()
	subj.auth.Properties = NewBagFromMap(map[string]string{
		"aws:SourceIp": "10.0.0.50",
	})

	stmt := policy.Statement{
		Effect: policy.EFFECT_ALLOW,
		Condition: map[string]policy.ConditionValues{
			"IpAddress": {
				"aws:SourceIp": {"10.0.0.0/8"},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evalStatementMatchesCondition(subj, &stmt)
	}
}

func BenchmarkEvalStatement_ConditionArnLike(b *testing.B) {
	subj := createTestSubject()
	subj.auth.Properties = NewBagFromMap(map[string]string{
		"aws:SourceArn": "arn:aws:sns:us-east-1:123456789012:my-topic",
	})

	stmt := policy.Statement{
		Effect: policy.EFFECT_ALLOW,
		Condition: map[string]policy.ConditionValues{
			"ArnLike": {
				"aws:SourceArn": {"arn:aws:sns:*:123456789012:*"},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evalStatementMatchesCondition(subj, &stmt)
	}
}

// -------------------------------------------------------------------------------------------------
// Full Simulation Benchmarks
// -------------------------------------------------------------------------------------------------

func BenchmarkSimulate_Simple(b *testing.B) {
	sim, err := buildTestSimulator()
	if err != nil {
		b.Fatalf("failed to build simulator: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = sim.SimulateByArn(
			"arn:aws:iam::777583092761:role/RedRole",
			"s3:listbucket",
			"arn:aws:s3:::yams-magenta",
		)
	}
}

func BenchmarkSimulate_WithConditions(b *testing.B) {
	sim, err := buildTestSimulator()
	if err != nil {
		b.Fatalf("failed to build simulator: %v", err)
	}

	opts := NewOptions(
		WithSkipServiceAuthorizationValidation(),
		WithAdditionalProperties(map[string]string{
			"aws:RequestedRegion": "us-east-1",
			"aws:SourceIp":        "10.0.0.0",
		}),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = sim.SimulateByArnWithOptions(
			"arn:aws:iam::255082776537:role/SandwichRole",
			"s3:getobject",
			"arn:aws:s3:::banana-bucket-255082776537/yams.txt",
			opts,
		)
	}
}

func BenchmarkSimulate_CrossAccount(b *testing.B) {
	sim, err := buildTestSimulator()
	if err != nil {
		b.Fatalf("failed to build simulator: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = sim.SimulateByArn(
			"arn:aws:iam::777583092761:role/BlueRole",
			"sns:publish",
			"arn:aws:sns:us-east-1:213308312933:LemurTopic",
		)
	}
}

func BenchmarkSimulate_AssumeRole(b *testing.B) {
	sim, err := buildTestSimulator()
	if err != nil {
		b.Fatalf("failed to build simulator: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = sim.SimulateByArn(
			"arn:aws:iam::777583092761:role/GreenRole",
			"sts:assumerole",
			"arn:aws:iam::777583092761:role/MustardRole",
		)
	}
}

// -------------------------------------------------------------------------------------------------
// Batch/Product Simulation Benchmarks
// -------------------------------------------------------------------------------------------------

func BenchmarkWhichActions(b *testing.B) {
	sim, err := buildTestSimulator()
	if err != nil {
		b.Fatalf("failed to build simulator: %v", err)
	}

	opts := NewOptions(WithSkipServiceAuthorizationValidation())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = sim.WhichActions(
			"arn:aws:iam::777583092761:role/BlueRole",
			"arn:aws:s3:::yams-cyan",
			opts,
		)
	}
}

func BenchmarkWhichResources(b *testing.B) {
	sim, err := buildTestSimulator()
	if err != nil {
		b.Fatalf("failed to build simulator: %v", err)
	}

	opts := NewOptions(WithSkipServiceAuthorizationValidation())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = sim.WhichResources(
			"arn:aws:iam::777583092761:role/BlueRole",
			"s3:listbucket",
			opts,
		)
	}
}

// -------------------------------------------------------------------------------------------------
// Large-scale Benchmarks (uses notes/resources.json.gz if available)
// -------------------------------------------------------------------------------------------------

func loadLargeScaleUniverse(b *testing.B) *entities.Universe {
	file, err := os.Open("../../notes/resources.json.gz")
	if err != nil {
		b.Skip("notes/resources.json.gz not available, skipping large-scale benchmark")
		return nil
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		b.Fatalf("failed to create gzip reader: %v", err)
	}
	defer gzReader.Close()

	loader := awsconfig.NewLoader()
	if err := loader.LoadJson(gzReader); err != nil {
		b.Fatalf("failed to load JSON: %v", err)
	}

	return loader.Universe()
}

func BenchmarkLargeScale_PrincipalFreeze(b *testing.B) {
	uv := loadLargeScaleUniverse(b)
	if uv == nil {
		return
	}

	principals := uv.PrincipalArns()
	if len(principals) == 0 {
		b.Skip("no principals in universe")
		return
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, arn := range principals[:min(100, len(principals))] {
			if p, ok := uv.Principal(arn); ok {
				_, _ = p.Freeze()
			}
		}
	}
}

func BenchmarkLargeScale_ResourceLookup(b *testing.B) {
	uv := loadLargeScaleUniverse(b)
	if uv == nil {
		return
	}

	resources := uv.ResourceArns()
	if len(resources) == 0 {
		b.Skip("no resources in universe")
		return
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, arn := range resources[:min(1000, len(resources))] {
			uv.Resource(arn)
		}
	}
}

// -------------------------------------------------------------------------------------------------
// Condition Operator Benchmarks
// -------------------------------------------------------------------------------------------------

func BenchmarkCondition_ResolveOperator(b *testing.B) {
	operators := []string{
		"StringEquals",
		"StringLike",
		"ArnLike",
		"IpAddress",
		"NumericEquals",
		"DateLessThan",
		"ForAllValues:StringEquals",
		"StringEqualsIfExists",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, op := range operators {
			ResolveConditionEvaluator(op)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
