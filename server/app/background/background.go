package background

import (
	"context"
)

type BackgroundTask interface {
	Start(ctx context.Context)
}
