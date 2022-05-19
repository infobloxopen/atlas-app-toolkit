package bloxid

func WithSchemer(fnOpt GenerateV0Opts) func(o *V0Options) {
	return func(o *V0Options) {
		fnOpt(o)
	}
}

// Schemer defines the interface required for encoding/decoding different schemes
type Schemer interface {
	FromEntityID(opts *V0Options) (scheme string, decoded string, encoded string, err error)
}
