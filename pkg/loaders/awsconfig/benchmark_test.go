package awsconfig

import (
	"compress/gzip"
	"os"
	"runtime"
	"testing"
)

const largeBenchFile = "../../../notes/resources.json.gz"

func BenchmarkLoadJson_LargeUniverse(b *testing.B) {
	// Check file exists
	if _, err := os.Stat(largeBenchFile); os.IsNotExist(err) {
		b.Skip("Large benchmark file not found:", largeBenchFile)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()

		// Open and decompress file
		f, err := os.Open(largeBenchFile)
		if err != nil {
			b.Fatal(err)
		}

		gz, err := gzip.NewReader(f)
		if err != nil {
			f.Close()
			b.Fatal(err)
		}

		l := NewLoader()

		// Start timing
		b.StartTimer()
		err = l.LoadJson(gz)
		b.StopTimer()

		if err != nil {
			b.Fatal(err)
		}

		// Report allocations
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		b.ReportMetric(float64(m.Alloc)/(1024*1024), "MB-in-use")
		b.ReportMetric(float64(l.Universe().Size()), "entities")

		gz.Close()
		f.Close()
	}
}
