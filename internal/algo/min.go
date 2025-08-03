package algo

import "math"

func Min(nums ...int) int {
	if len(nums) == 0 {
		return 0
	}

	min := math.MaxInt
	for _, v := range nums {
		if v < min {
			min = v
		}
	}

	return min
}
