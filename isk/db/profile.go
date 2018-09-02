package db

// Profile describes someone setup to receive ISK
type Profile struct {
	// CharacterID recipient
	CharacterID int32

	// Private is if we should hide this profile for everyone else
	Private bool

	// Passphrase can be set by the profile owner for minimal security
	Passphrase string
}
