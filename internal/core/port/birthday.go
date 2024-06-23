package port

import (
	"context"
	"sync"
)

//go:generate mockgen -source=./birthday.go -destination=mock/birthday.go -package=mock

type Birthday interface {
	BirthdayNotify(ctx context.Context, wg *sync.WaitGroup)
}
