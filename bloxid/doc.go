// Bloxid implements typed guids for identifying resource
// objects globally in the system. Resources are not specific
// to services/applications but must contain an entity type.
//
// Typed guids have the advantage of being globally unique,
// easily readable, and strongly typed for authorization and
// logging. The trailing characters provide sufficient entropy
// to make each resource universally unique.
//
// Bloxid package provides methods for generating and parsing
// versioned typed guids.
package bloxid
