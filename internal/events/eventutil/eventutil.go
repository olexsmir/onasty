package eventutil

import "context"

func Request[T comparable, R comparable](ctx context.Context, input T) (R, error) {
}
