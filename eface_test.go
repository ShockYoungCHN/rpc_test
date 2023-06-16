package main

import "testing"

func BenchmarkEfaceScalar(b *testing.B) {
	var _ uint32
	b.Run("uint32", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = uint32(i)
		}
	})
	var a interface{}
	b.Run("eface32", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			a = uint32(i)
			// prevent compiler from optimizing away the conversion
			// stop the timer
			b.StopTimer()
			_ = a.(uint32)
			b.StartTimer()
		}
	})
}
