package gmap

func SafeAssign[K comparable, V any](m map[K]V, key K, value V) map[K]V {
	if m == nil {
		return map[K]V{key: value}
	}

	m[key] = value
	return m
}
