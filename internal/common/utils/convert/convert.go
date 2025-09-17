package convert

func Setint64(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

func Setfloat64(f *float64) float64 {
	if f == nil {
		return 0.0
	}
	return *f
}
