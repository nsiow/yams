package condition

// This const block holds string constants corresponding to AWS condition operators
const (
	// ---------------------------------------------------------------------------------------------
	// String Functions
	// ---------------------------------------------------------------------------------------------

	StringEquals              = "StringEquals"
	StringNotEquals           = "StringNotEquals"
	StringEqualsIgnoreCase    = "StringEqualsIgnoreCase"
	StringNotEqualsIgnoreCase = "StringNotEqualsIgnoreCase"
	StringLike                = "StringLike"
	StringNotLike             = "StringNotLike"
	StringLikeIgnoreCase      = "StringLikeIgnoreCase"
	StringNotLikeIgnoreCase   = "StringNotLikeIgnoreCase"

	// ---------------------------------------------------------------------------------------------
	// Numeric Functions
	// ---------------------------------------------------------------------------------------------

	NumericEquals            = "NumericEquals"
	NumericNotEquals         = "NumericNotEquals"
	NumericLessThan          = "NumericLessThan"
	NumericLessThanEquals    = "NumericLessThanEquals"
	NumericGreaterThan       = "NumericGreaterThan"
	NumericGreaterThanEquals = "NumericGreaterThanEquals"

	// ---------------------------------------------------------------------------------------------
	// Date Functions
	// ---------------------------------------------------------------------------------------------

	DateEquals            = "DateEquals"
	DateNotEquals         = "DateNotEquals"
	DateLessThan          = "DateLessThan"
	DateLessThanEquals    = "DateLessThanEquals"
	DateGreaterThan       = "DateGreaterThan"
	DateGreaterThanEquals = "DateGreaterThanEquals"

	// ---------------------------------------------------------------------------------------------
	// Bool Functions
	// ---------------------------------------------------------------------------------------------

	Bool = "Bool"

	// ---------------------------------------------------------------------------------------------
	// Binary Functions
	// ---------------------------------------------------------------------------------------------

	BinaryEquals = "BinaryEquals"

	// ---------------------------------------------------------------------------------------------
	// IP Functions
	// ---------------------------------------------------------------------------------------------

	IpAddress    = "IpAddress"
	NotIpAddress = "NotIpAddress"

	// ---------------------------------------------------------------------------------------------
	// Arn Functions
	// ---------------------------------------------------------------------------------------------

	// Arn functions
	ArnEquals    = "ArnEquals"
	ArnNotEquals = "ArnNotEquals"
	ArnLike      = "ArnLike"
	ArnNotLike   = "ArnNotLike"

	// TODO(nsiow) add Null operator
	// https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_condition_operators.html#Conditions_Null
)
