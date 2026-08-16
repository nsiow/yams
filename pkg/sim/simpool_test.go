package sim

import (
	"context"
	"runtime"
	"testing"
	"time"
)

func TestPoolDefaults(t *testing.T) {
	t.Setenv("YAMS_SIM_NUM_WORKERS", "")
	t.Setenv("YAMS_SIM_BATCH_SIZE", "")
	t.Setenv("YAMS_SIM_TIMEOUT", "")

	pool := NewPool(context.Background(), nil)
	if got, want := pool.NumWorkers(), runtime.NumCPU(); got != want {
		t.Errorf("NumWorkers() = %d, want %d", got, want)
	}
	if got, want := pool.BatchSize(), defaultBatchSize; got != want {
		t.Errorf("BatchSize() = %d, want %d", got, want)
	}
	if got, want := pool.Timeout(), time.Minute; got != want {
		t.Errorf("Timeout() = %s, want %s", got, want)
	}
}

func TestPoolConfigurationFromEnvironment(t *testing.T) {
	t.Setenv("YAMS_SIM_NUM_WORKERS", "8")
	t.Setenv("YAMS_SIM_BATCH_SIZE", "512")
	t.Setenv("YAMS_SIM_TIMEOUT", "120")

	pool := NewPool(context.Background(), nil)
	if got, want := pool.NumWorkers(), 8; got != want {
		t.Errorf("NumWorkers() = %d, want %d", got, want)
	}
	if got, want := pool.BatchSize(), 512; got != want {
		t.Errorf("BatchSize() = %d, want %d", got, want)
	}
	if got, want := pool.Timeout(), 2*time.Minute; got != want {
		t.Errorf("Timeout() = %s, want %s", got, want)
	}
}

func TestPositiveEnvInt(t *testing.T) {
	for _, test := range []struct {
		name  string
		value string
		want  int
	}{
		{name: "positive", value: "12", want: 12},
		{name: "zero", value: "0", want: 7},
		{name: "negative", value: "-1", want: 7},
		{name: "invalid", value: "invalid", want: 7},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("YAMS_TEST_POSITIVE_INT", test.value)
			if got := positiveEnvInt("YAMS_TEST_POSITIVE_INT", 7); got != test.want {
				t.Fatalf("positiveEnvInt() = %d, want %d", got, test.want)
			}
		})
	}
}
