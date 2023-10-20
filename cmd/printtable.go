package cmd

import (
	"fmt"
)

func PrintTable(headers []string, data [][]string) {
	// Calculate the maximum width of each column
	widths := make([]int, len(headers))
	for i, header := range headers {
		widths[i] = len(header)
	}
	for _, row := range data {
		for i, cell := range row {
			if len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print the column headers
	for i, header := range headers {
		fmt.Printf(Blue+"%-*s"+Reset, widths[i]+2, header)
	}
	fmt.Println()

	// Print the data rows
	for _, row := range data {
		for i, cell := range row {
			fmt.Printf("%-*s", widths[i]+2, cell)
		}
		fmt.Println()
	}
}
