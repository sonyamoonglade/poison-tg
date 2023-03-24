package addr

func Int[V int | int8 | int16 | int32 | int64](v V) *V {
	return &v
}
