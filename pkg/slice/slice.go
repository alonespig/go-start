package slice

// Map 对切片每个元素应用 fn，返回新切片
func Map[T, R any](items []T, fn func(T) R) []R {
	out := make([]R, 0, len(items))
	for _, it := range items {
		out = append(out, fn(it))
	}
	return out
}
