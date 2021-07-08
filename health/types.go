package health

import "context"

// Check
type Check func() error

type CheckContext func(ctx context.Context) error
