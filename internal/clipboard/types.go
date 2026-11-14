package clipboard

import (
	"context"

	"golang.design/x/clipboard"
)

type clipboardWatchFunc func(
	ctx context.Context,
	opts ...clipboard.Option,
) <-chan clipboard.Data

type clipboardWriteFunc func(
	ctx context.Context,
	format clipboard.Format,
	data []byte,
	opts ...clipboard.Option,
) (<-chan struct{}, error)
