package apaas

// cloneMap creates a shallow copy of the provided map to avoid unexpected mutations.
func cloneMap(m map[string]any) map[string]any {
	if m == nil {
		return map[string]any{}
	}

	c := make(map[string]any, len(m))
	for k, v := range m {
		c[k] = v
	}
	return c
}
