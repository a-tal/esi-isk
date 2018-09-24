package cx

// Key describes any context key used in ESI ISK
type Key string

const (
	/* -- Global Keys -- */

	// Opts is our global server runtime (*cx.Options)
	Opts = Key("Opts")

	// Provider holds the oidc provider
	Provider = Key("Provider")

	// Verifier verifies JWTs
	Verifier = Key("Verifier")

	// DB is our pg connection (*sqlx.DB)
	DB = Key("DB")

	// Cache is our httpCache object
	Cache = Key("Cache")

	// Statements is our map of prepared statements (map[Key]sqlx.Stmt)
	Statements = Key("Statements")

	// StateStore is our in-memory auth state store (*api.StateStore)
	StateStore = Key("StateStore")

	// SSOClient is XXX HACK REMOVE ME
	SSOClient = Key("SSOClient")

	// Client is the goesi client
	Client = Key("Client")

	// HTTPClient is the http.Client the goesi API Client is using
	HTTPClient = Key("HTTPClient")

	// Authenticator is the global goesi SSO authenticator
	Authenticator = Key("Authenticator")

	/* -- API Statements -- */

	// StmtTopReceived pulls the top character_id and receiver totals
	StmtTopReceived = Key("TestStatement")

	// StmtTopDonated pulls the top character_id and donation totals
	StmtTopDonated = Key("StmtTopDonated")

	// StmtCharDetails pulls details for a specific character
	StmtCharDetails = Key("StmtCharDetails")

	// StmtCharDonations pulls the donations to a character
	StmtCharDonations = Key("StmtCharDonations")

	// StmtCharContracts pulls the contracts to a character
	StmtCharContracts = Key("StmtCharContracts")

	// StmtCharDonated pulls the donations from a character
	StmtCharDonated = Key("StmtCharDonated")

	// StmtCharContracted pulls the contracts from a character
	StmtCharContracted = Key("StmtCharContracted")

	// StmtContractItems pulls the items for a contract
	StmtContractItems = Key("StmtContractItems")

	// StmtUserChars pulls the character ids for a userID
	StmtUserChars = Key("StmtUserChars")

	// StmtCreateUser creates a new user with a paired character
	StmtCreateUser = Key("StmtCreateUser")

	// StmtGetUser pulls a user by ID
	StmtGetUser = Key("StmtGetUser")

	// StmtGetUsers pulls users needing an update (up to 100)
	StmtGetUsers = Key("StmtGetUsers")

	// StmtGetNullUsers pulls users with null last processed timestamps
	StmtGetNullUsers = Key("StmtGetNullUsers")

	// StmtUpdateUser updates a user's character (auth updates)
	StmtUpdateUser = Key("StmtUpdateUser")

	// StmtDeleteUser deletes a user
	StmtDeleteUser = Key("StmtDeleteUser")

	// StmtAddDonation inserts a donation into the donations table
	StmtAddDonation = Key("StmtAddDonation")

	// StmtGetName returns the name for an ID
	StmtGetName = Key("StmtGetName")

	// StmtNewName creates a new mapping of ID<->name
	StmtNewName = Key("StmtNewName")

	// StmtUpdateName updates a mapping of ID<->name
	StmtUpdateName = Key("StmtUpdateName")

	// StmtCreateCharacter creates a new character
	StmtCreateCharacter = Key("StmtCreateCharacter")

	// StmtUpdateCharacter updates a known character
	StmtUpdateCharacter = Key("StmtUpdateCharacter")

	// StmtAddContract creates a new contract
	StmtAddContract = Key("StmtAddContract")

	// StmtAddContractItems creates a new contract item row
	StmtAddContractItems = Key("StmtAddContractItems")

	// StmtCharStandingISK queries donations towards standings
	StmtCharStandingISK = Key("StmtCharStandingISK")
)
