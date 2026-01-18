package condition

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

	// ---------------------------------------------------------------------------------------------
	// Null Function
	// ---------------------------------------------------------------------------------------------

	// Null checks whether a condition key is present in the request context.
	// "true" matches when the key is absent; "false" matches when the key is present.
	Null = "Null"
)
