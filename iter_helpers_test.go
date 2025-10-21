package threadsafe

import "iter"

// collectSeq exhausts seq into a slice.
func collectSeq[T any](seq iter.Seq[T]) []T {
	var out []T
	for v := range seq {
		out = append(out, v)
	}
	return out
}

// collectSeq2 exhausts seq into key/value slices.
func collectSeq2[K any, V any](seq iter.Seq2[K, V]) ([]K, []V) {
	var keys []K
	var values []V
	for k, v := range seq {
		keys = append(keys, k)
		values = append(values, v)
	}
	return keys, values
}
