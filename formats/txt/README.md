# Text Format

A utility packge for writing text files. For cases where you want to avoid fmt.FprintX, offers some speed advantage.

Simple benchmark provided presents a 69(nice)% runtime improvement.

```
goos: windows
goarch: amd64
pkg: github.com/EliCDavis/polyform/formats/txt
cpu: 13th Gen Intel(R) Core(TM) i7-13800H
BenchmarkFormat_Fprintf-20               	 2945271	       405.8 ns/op	     119 B/op	       0 allocs/op
BenchmarkFormat_TextWriter-20            	 9266608	       124.3 ns/op	      81 B/op	       0 allocs/op
```