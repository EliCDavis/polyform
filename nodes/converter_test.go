package nodes_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
)

// Golden represents a test case.
type Golden struct {
	name        string
	trimPrefix  string
	lineComment bool
	input       string // input; the package clause is provided when running the test.
	output      string // expected output.
}

var golden = []Golden{
	{"day", "", false, day_in, day_out},
	// {"offset", "", false, offset_in, offset_out},
	// {"gap", "", false, gap_in, gap_out},
	// {"num", "", false, num_in, num_out},
	// {"unum", "", false, unum_in, unum_out},
	// {"unumpos", "", false, unumpos_in, unumpos_out},
	// {"prime", "", false, prime_in, prime_out},
	// {"prefix", "Type", false, prefix_in, prefix_out},
	// {"tokens", "", true, tokens_in, tokens_out},
}

const day_in = `type Day int
const (
	Monday Day = iota
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
	Sunday
)

func Ffuucckkk(v string) (int, error) {
	return 0, nil
}
`

const day_out = `func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Monday-0]
	_ = x[Tuesday-1]
	_ = x[Wednesday-2]
	_ = x[Thursday-3]
	_ = x[Friday-4]
	_ = x[Saturday-5]
	_ = x[Sunday-6]
}

const _Day_name = "MondayTuesdayWednesdayThursdayFridaySaturdaySunday"

var _Day_index = [...]uint8{0, 6, 13, 22, 30, 36, 44, 50}

func (i Day) String() string {
	if i < 0 || i >= Day(len(_Day_index)-1) {
		return "Day(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Day_name[_Day_index[i]:_Day_index[i+1]]
}
`

func TestGolden(t *testing.T) {
	// testenv.NeedsTool(t, "go")

	dir := t.TempDir()
	for _, test := range golden {
		test := test
		t.Run(test.name, func(t *testing.T) {
			// g := Generator{
			// 	trimPrefix:  test.trimPrefix,
			// 	lineComment: test.lineComment,
			// 	logf:        t.Logf,
			// }
			input := "package test\n" + test.input
			file := test.name + ".go"
			absFile := filepath.Join(dir, file)
			err := os.WriteFile(absFile, []byte(input), 0644)
			if err != nil {
				t.Fatal(err)
			}

			// g.parsePackage([]string{absFile}, nil)
			// Extract the name and type of the constant from the first line.
			tokens := strings.SplitN(test.input, " ", 3)
			if len(tokens) != 3 {
				t.Fatalf("%s: need type declaration on first line", test.name)
			}

			out := nodes.Convert([]string{absFile})
			assert.Equal(t, test.output, out)
			// g.generate(tokens[1])
			// got := string(g.format())
			// if got != test.output {
			// 	t.Errorf("%s: got(%d)\n====\n%q====\nexpected(%d)\n====%q", test.name, len(got), got, len(test.output), test.output)
			// }
		})
	}
}
