package crongo

import "context"

type Tasker interface {
	Run(ctx context.Context) error
}
