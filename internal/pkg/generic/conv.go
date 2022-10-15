package generic

import "strconv"

func Itoa[typ UnderlyingInteger](i typ) string {
	return strconv.Itoa(int(i))
}

func Btoi(b bool) int {
	var result int
	if b {
		result = 1
	}
	return result
}
