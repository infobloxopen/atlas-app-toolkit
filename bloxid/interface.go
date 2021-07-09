package bloxid

// ID implements the interface for parsing a resource identifier
type ID interface {
	// String returns the complete resource ID
	String() string
	// ShortID returns a shortened ID that will be locally unique
	ShortID() string
	// Version returns a serialized representation of the ID version
	// ie. `V0`
	Version() string // V0
	// Type returns entity type ie. `host`
	Type() string
	// Realm is optional and returns the cloud realm that
	// the resource is found in ie. `us-com-1`, `eu-com-1`, ...
	Realm() string
}
