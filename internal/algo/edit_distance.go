package algo

func EditDistance(word1, word2 string) int {
	rows := len(word1)
	cols := len(word2)

	table := make([][]int, rows+1)
	for k := range rows + 1 {
		table[k] = make([]int, cols+1)
	}

	for i := cols; i >= 0; i-- {
		table[rows][i] = cols - i
	}
	for i := rows; i >= 0; i-- {
		table[i][cols] = rows - i
	}

	for r := rows - 1; r >= 0; r-- {
		for c := cols - 1; c >= 0; c-- {
			if word1[r] == word2[c] {
				table[r][c] = table[r+1][c+1]
				continue
			}
			table[r][c] = Min(
				table[r+1][c],
				table[r][c+1],
				table[r+1][c+1],
			) + 1
		}
	}

	return table[0][0]
}
