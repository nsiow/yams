package entities

// Account defines the general shape of an AWS account
type Account struct {
	// uv is a reverse pointer back to the containing universe
	uv *Universe `json:"-"`

	// AccountId refers to the 12-digit ID of this AWS account
	Id string

	// OrgId refers to the ID of the AWS Organizations org where the Account resides
	OrgId string

	// OrgPaths refers to the collection of org-paths containing the account
	// TODO(nsiow) implement this in the org crawler
	OrgPaths []string

	// OrgNodes refers to the path from the organizations root to the account itself
	//
	// It is INCLUSIVE of the account itself, which is to say that [OrgNodes] will include an OrgNode
	// with Type=ACCOUNT and Id=Account.id
	OrgNodes []OrgNode
}
