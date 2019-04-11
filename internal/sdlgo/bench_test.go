package sdlgo_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"gitlabe1.ext.net.nokia.com/ric_dev/sdlgo"
)

type singleBenchmark struct {
	key       string
	keySize   int
	valueSize int
}

type multiBenchmark struct {
	keyBase   string
	keyCount  int
	keySize   int
	valueSize int
}

func (bm singleBenchmark) String(oper string) string {
	return fmt.Sprintf("op = %s key=%d value=%d", oper, bm.keySize, bm.valueSize)
}

func (bm multiBenchmark) String(oper string) string {
	return fmt.Sprintf("op = %s keycnt=%d key=%d value=%d", oper, bm.keyCount, bm.keySize, bm.valueSize)
}
func BenchmarkSet(b *testing.B) {
	benchmarks := []singleBenchmark{
		{"a", 10, 64},
		{"b", 10, 1024},
		{"c", 10, 64 * 1024},
		{"d", 10, 1024 * 1024},
		{"e", 10, 10 * 1024 * 1024},

		{"f", 100, 64},
		{"g", 100, 1024},
		{"h", 100, 64 * 1024},
		{"i", 100, 1024 * 1024},
		{"j", 100, 10 * 1024 * 1024},
	}

	for _, bm := range benchmarks {
		b.Run(bm.String("set"), func(b *testing.B) {
			key := strings.Repeat(bm.key, bm.keySize)
			value := strings.Repeat("1", bm.valueSize)
			sdl := sdlgo.Create("namespace")

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					err := sdl.Set(key, value)
					if err != nil {
						b.Fatal(err)
					}
				}
			})
		})
	}
}

func BenchmarkGet(b *testing.B) {
	benchmarks := []singleBenchmark{
		{"a", 10, 64},
		{"b", 10, 1024},
		{"c", 10, 64 * 1024},
		{"d", 10, 1024 * 1024},
		{"e", 10, 10 * 1024 * 1024},

		{"f", 100, 64},
		{"g", 100, 1024},
		{"h", 100, 64 * 1024},
		{"i", 100, 1024 * 1024},
		{"j", 100, 10 * 1024 * 1024},
	}

	for _, bm := range benchmarks {
		b.Run(bm.String("Get"), func(b *testing.B) {
			key := strings.Repeat(bm.key, bm.keySize)
			value := strings.Repeat("1", bm.valueSize)
			sdl := sdlgo.Create("namespace")
			if err := sdl.Set(key, value); err != nil {
				b.Fatal(err)
			}
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					_, err := sdl.Get([]string{key})
					if err != nil {
						b.Fatal(err)
					}
				}
			})
		})
	}
}

func BenchmarkMultiSet(b *testing.B) {
	benchmarks := []multiBenchmark{
		{"a", 2, 10, 64},
		{"b", 10, 10, 64},
		{"c", 100, 10, 64},
		{"d", 1000, 10, 64},
		{"e", 5000, 10, 64},

		{"f", 2, 100, 64},
		{"g", 10, 100, 64},
		{"h", 100, 100, 64},
		{"i", 1000, 100, 64},
		{"j", 5000, 100, 64},
	}

	for _, bm := range benchmarks {
		b.Run(bm.String("mset"), func(b *testing.B) {
			sdl := sdlgo.Create("namespace")
			value := strings.Repeat("1", bm.valueSize)
			keyVals := make([]string, 0)
			for i := 0; i < bm.keyCount; i++ {
				key := strings.Repeat(bm.keyBase+strconv.Itoa(i), bm.keySize)
				keyVals = append(keyVals, key, value)
			}
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					err := sdl.Set(keyVals)
					if err != nil {
						b.Fatal(err)
					}
				}
			})
		})
	}
}

func BenchmarkMultiGet(b *testing.B) {
	benchmarks := []multiBenchmark{
		{"a", 2, 10, 64},
		{"b", 10, 10, 64},
		{"c", 100, 10, 64},
		{"d", 1000, 10, 64},
		{"e", 5000, 10, 64},

		{"f", 2, 100, 64},
		{"g", 10, 100, 64},
		{"h", 100, 100, 64},
		{"i", 1000, 100, 64},
		{"j", 5000, 100, 64},
	}

	for _, bm := range benchmarks {
		b.Run(bm.String("gset"), func(b *testing.B) {
			sdl := sdlgo.Create("namespace")
			keyVals := make([]string, 0)
			for i := 0; i < bm.keyCount; i++ {
				key := strings.Repeat(bm.keyBase+strconv.Itoa(i), bm.keySize)
				keyVals = append(keyVals, key)
			}
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					_, err := sdl.Get(keyVals)
					if err != nil {
						b.Fatal(err)
					}
				}
			})
		})
	}
}
