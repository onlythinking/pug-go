package logging_test

import (
	"context"
	"testing"

	"github.com/onlythinking/pug-go/pkg/logging"
)

func TestNewLogger(t *testing.T) {
	t.Parallel()

	logger := logging.NewLogger(true)
	if logger == nil {
		t.Fatal("expected logger to never be nil")
	}
}

func TestDefaultLogger(t *testing.T) {
	t.Parallel()

	logger1 := logging.DefaultLogger()
	if logger1 == nil {
		t.Fatal("expected logger to never be nil")
	}

	logger2 := logging.DefaultLogger()
	if logger2 == nil {
		t.Fatal("expected logger to never be nil")
	}

	if logger1 != logger2 {
		t.Errorf("expected %#v to be %#v", logger1, logger2)
	}
}

func TestContext(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger1 := logging.FromContext(ctx)
	if logger1 == nil {
		t.Fatal("expected logger to never be nil")
	}

	ctx = logging.WithLogger(ctx, logger1)

	logger2 := logging.FromContext(ctx)
	if logger1 != logger2 {
		t.Errorf("expected %#v to be %#v", logger1, logger2)
	}
}
