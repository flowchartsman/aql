package fmtmsg

import (
	"github.com/flowchartsman/aql/parser/ast"
)

func truncate(pd *printData, maxLen int) {
	// TODO: line-awareness
	// don't trim if no maxlen, pos is zero or query is below maxlen
	if maxLen <= 0 || pd.pos == ast.NoPosition() || len(pd.query) <= maxLen {
		return
	}
	pd.lo, pd.ro = findCenterSubstring(pd.query, maxLen, pd.pos.Offset, pd.pos.Offset+pd.pos.Len)
}

func findCenterSubstring(r []rune, windowSize int, startOffset int, endOffset int) (int, int) {
	strLen := len(r)
	substringLen := endOffset - startOffset + 1

	if windowSize >= strLen {
		// If the window size is greater than or equal to the string length,
		// the entire string is the substring.
		return 0, strLen
	}

	if substringLen >= windowSize {
		// If the specified substring length is greater than or equal to
		// the window size, return the substring truncated on the right to windowSize
		return startOffset, startOffset + windowSize
	}

	// Calculate the available space on both sides of the substring.
	leftSpace := startOffset
	rightSpace := strLen - endOffset - 1

	// Calculate the maximum possible size of the new substring within the window.
	maxSubstringSize := min(windowSize, substringLen+min(leftSpace, rightSpace))

	// Calculate the left and right offsets for the new substring.
	centerOffset := (startOffset + endOffset) / 2
	halfMaxSubstringSize := maxSubstringSize / 2

	leftOffset := centerOffset - halfMaxSubstringSize
	rightOffset := leftOffset + maxSubstringSize - 1

	// Adjust the offsets if they are too close to the beginning or end of the string.
	if leftOffset < 0 {
		rightOffset -= leftOffset
		leftOffset = 0
	}

	if rightOffset >= strLen {
		leftOffset -= (rightOffset - strLen + 1)
		rightOffset = strLen - 1
	}

	// Adjust the offsets if the new substring is smaller than the window size.
	if rightOffset-leftOffset+1 < windowSize {
		leftOffset = centerOffset - (windowSize-1)/2
		rightOffset = leftOffset + windowSize

		if leftOffset < 0 {
			leftOffset = 0
			rightOffset = min(windowSize, strLen)
		}

		if rightOffset >= strLen {
			rightOffset = strLen
			leftOffset = max(rightOffset-(windowSize), 0)
		}
	}

	// Adjust the right offset to match the window size.
	if rightOffset-leftOffset+1 > windowSize {
		rightOffset = leftOffset + windowSize
	}

	return leftOffset, rightOffset
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
