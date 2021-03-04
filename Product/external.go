package sitescraper

// bytestoString ...
func bytesToString(data []byte) string {
	return string(data[:])
}

// Min returns the smaller of x or y.
func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
