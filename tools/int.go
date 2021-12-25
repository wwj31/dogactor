package tools

func Int32Merge(h, l int16) (id int32) {
	return int32(uint32(h)<<16 | uint32(l))
}

func Int32Split(id int32) (h, l int16) {
	return int16(uint32(id) >> 16), int16(uint16(id))
}

func Maxi64(x, y int64) int64 {
	if x > y {
		return x
	}
	return y
}
func Maxi32(x, y int32) int32 {
	if x > y {
		return x
	}
	return y
}

func Mini64(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}
