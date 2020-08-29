package level

import (
	"context"
	"testing"
)

func TestInfo(t *testing.T) {
	ctx := context.Background()
	Info(ctx).F("hello", "world")
}
