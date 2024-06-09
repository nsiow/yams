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
	NumericGreaterThanEquals = "NumericGreaterThanEqual"
)
