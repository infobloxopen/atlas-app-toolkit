package bloxid

// ID implements the interface for parsing a resource identifier
type ID interface {
	// String returns the complete resource ID
	String() string
	// Version returns a serialized representation of the ID version
	// ie. `V0`
	Version() string // V0
	// Domain returns entity domain ie. `infra`
	Domain() string
	// Type returns entity type ie. `host`
	Type() string
	// Realm is optional and returns the cloud realm that
	// the resource is found in ie. `us-com-1`, `eu-com-1`, ...
	Realm() string
	// EncodedID returns the unique id in encoded format
	EncodedID() string
	// DecodedID returns the unique id in decoded format
	DecodedID() string
	// Return the id scheme
	Scheme() string
}
