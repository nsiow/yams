package condition

import "testing"

func TestConditionOperatorConstants(t *testing.T) {
	// Verify string operators
	if StringEquals != "StringEquals" {
		t.Errorf("StringEquals expected 'StringEquals', got '%s'", StringEquals)
	}
	if StringNotEquals != "StringNotEquals" {
		t.Errorf("StringNotEquals expected 'StringNotEquals', got '%s'", StringNotEquals)
	}
	if StringEqualsIgnoreCase != "StringEqualsIgnoreCase" {
		t.Errorf("StringEqualsIgnoreCase expected 'StringEqualsIgnoreCase', got '%s'", StringEqualsIgnoreCase)
	}
	if StringNotEqualsIgnoreCase != "StringNotEqualsIgnoreCase" {
		t.Errorf("StringNotEqualsIgnoreCase expected 'StringNotEqualsIgnoreCase', got '%s'", StringNotEqualsIgnoreCase)
	}
	if StringLike != "StringLike" {
		t.Errorf("StringLike expected 'StringLike', got '%s'", StringLike)
	}
	if StringNotLike != "StringNotLike" {
		t.Errorf("StringNotLike expected 'StringNotLike', got '%s'", StringNotLike)
	}

	// Verify numeric operators
	if NumericEquals != "NumericEquals" {
		t.Errorf("NumericEquals expected 'NumericEquals', got '%s'", NumericEquals)
	}
	if NumericNotEquals != "NumericNotEquals" {
		t.Errorf("NumericNotEquals expected 'NumericNotEquals', got '%s'", NumericNotEquals)
	}
	if NumericLessThan != "NumericLessThan" {
		t.Errorf("NumericLessThan expected 'NumericLessThan', got '%s'", NumericLessThan)
	}
	if NumericLessThanEquals != "NumericLessThanEquals" {
		t.Errorf("NumericLessThanEquals expected 'NumericLessThanEquals', got '%s'", NumericLessThanEquals)
	}
	if NumericGreaterThan != "NumericGreaterThan" {
		t.Errorf("NumericGreaterThan expected 'NumericGreaterThan', got '%s'", NumericGreaterThan)
	}
	if NumericGreaterThanEquals != "NumericGreaterThanEquals" {
		t.Errorf("NumericGreaterThanEquals expected 'NumericGreaterThanEquals', got '%s'", NumericGreaterThanEquals)
	}

	// Verify date operators
	if DateEquals != "DateEquals" {
		t.Errorf("DateEquals expected 'DateEquals', got '%s'", DateEquals)
	}
	if DateNotEquals != "DateNotEquals" {
		t.Errorf("DateNotEquals expected 'DateNotEquals', got '%s'", DateNotEquals)
	}
	if DateLessThan != "DateLessThan" {
		t.Errorf("DateLessThan expected 'DateLessThan', got '%s'", DateLessThan)
	}
	if DateLessThanEquals != "DateLessThanEquals" {
		t.Errorf("DateLessThanEquals expected 'DateLessThanEquals', got '%s'", DateLessThanEquals)
	}
	if DateGreaterThan != "DateGreaterThan" {
		t.Errorf("DateGreaterThan expected 'DateGreaterThan', got '%s'", DateGreaterThan)
	}
	if DateGreaterThanEquals != "DateGreaterThanEquals" {
		t.Errorf("DateGreaterThanEquals expected 'DateGreaterThanEquals', got '%s'", DateGreaterThanEquals)
	}

	// Verify other operators
	if Bool != "Bool" {
		t.Errorf("Bool expected 'Bool', got '%s'", Bool)
	}
	if BinaryEquals != "BinaryEquals" {
		t.Errorf("BinaryEquals expected 'BinaryEquals', got '%s'", BinaryEquals)
	}
	if IpAddress != "IpAddress" {
		t.Errorf("IpAddress expected 'IpAddress', got '%s'", IpAddress)
	}
	if NotIpAddress != "NotIpAddress" {
		t.Errorf("NotIpAddress expected 'NotIpAddress', got '%s'", NotIpAddress)
	}

	// Verify ARN operators
	if ArnEquals != "ArnEquals" {
		t.Errorf("ArnEquals expected 'ArnEquals', got '%s'", ArnEquals)
	}
	if ArnNotEquals != "ArnNotEquals" {
		t.Errorf("ArnNotEquals expected 'ArnNotEquals', got '%s'", ArnNotEquals)
	}
	if ArnLike != "ArnLike" {
		t.Errorf("ArnLike expected 'ArnLike', got '%s'", ArnLike)
	}
	if ArnNotLike != "ArnNotLike" {
		t.Errorf("ArnNotLike expected 'ArnNotLike', got '%s'", ArnNotLike)
	}
}
