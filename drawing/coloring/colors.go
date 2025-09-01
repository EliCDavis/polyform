package coloring

func Black() WebColor {
	return WebColor{0, 0, 0, 1}
}

func Grey(b float64) WebColor {
	return WebColor{b, b, b, 1}
}

func White() WebColor {
	return WebColor{1, 1, 1, 1}
}

// func Red() WebColor {
// 	return WebColor{1, 0, 0, 1}
// }

// func Green() WebColor {
// 	return WebColor{0, 1, 0, 1}
// }

// func Blue() WebColor {
// 	return WebColor{0, 0, 1, 1}
// }
