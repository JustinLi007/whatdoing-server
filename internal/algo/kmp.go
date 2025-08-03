package algo

func Kmp(haystack string, needle string) int {
	hayLen := len(haystack)
	needleLen := len(needle)
	lps := make([]int, needleLen)

	needleIdx := 1
	prevLps := 0
	for needleIdx < needleLen {
		if needle[needleIdx] == needle[prevLps] {
			lps[needleIdx] = prevLps + 1
			needleIdx++
			prevLps++
		} else if prevLps == 0 {
			needleIdx++
		} else {
			prevLps = lps[prevLps-1]
		}
	}

	needleIdx = 0
	hayIdx := 0
	for hayIdx < hayLen {
		if haystack[hayIdx] == needle[needleIdx] {
			needleIdx++
			hayIdx++
		} else if needleIdx == 0 {
			hayIdx++
		} else {
			needleIdx = lps[needleIdx-1]
		}

		if needleIdx >= needleLen {
			return hayIdx - needleLen
		}
	}

	return -1
}
