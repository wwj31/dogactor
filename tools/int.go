package tools

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

func Max(x, y int) int {
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

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
