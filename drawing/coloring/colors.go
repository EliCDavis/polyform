package coloring

func Black() Color {
	return Color{0, 0, 0, 1}
}

func Grey(b float64) Color {
	return Color{b, b, b, 1}
}

func White() Color {
	return Color{1, 1, 1, 1}
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
