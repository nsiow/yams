package condition

// This const block holds string constants corresponding to AWS condition operators
// TODO(nsiow) add comments + references
const (
	// ------------------------------------------------------------------------------
	// String Functions
	// ------------------------------------------------------------------------------

	StringEquals              = "StringEquals"
	StringNotEquals           = "StringNotEquals"
	StringEqualsIgnoreCase    = "StringEqualsIgnoreCase"
	StringNotEqualsIgnoreCase = "StringNotEqualsIgnoreCase"
	StringLike                = "StringLike"
	StringNotLike             = "StringNotLike"
	StringLikeIgnoreCase      = "StringLikeIgnoreCase"
	StringNotLikeIgnoreCase   = "StringNotLikeIgnoreCase"

	// Numeric Functions
	NumericEquals            = "NumericEquals"
	NumericNotEquals         = "NumericNotEquals"
	NumericLessThan          = "NumericLessThan"
	NumericLessThanEquals    = "NumericLessThanEquals"
	NumericGreaterThan       = "NumericGreaterThan"
	NumericGreaterThanEquals = "NumericGreaterThanEquals"

	// Date Functions
	DateEquals            = "DateEquals"
	DateNotEquals         = "DateNotEquals"
	DateLessThan          = "DateLessThan"
	DateLessThanEquals    = "DateLessThanEquals"
	DateGreaterThan       = "DateGreaterThan"
	DateGreaterThanEquals = "DateGreaterThanEquals"

	// Bool Functions
	Bool = "Bool"

	// Binary Functions
	BinaryEquals = "BinaryEquals"

	// IP Address Functions
	IpAddress    = "IpAddress"
	NotIpAddress = "NotIpAddress"

	// Arn functions
	ArnEquals    = "ArnEquals"
	ArnNotEquals = "ArnNotEquals"
	ArnLike      = "ArnLike"
	ArnNotLike   = "ArnNotLike"
)
