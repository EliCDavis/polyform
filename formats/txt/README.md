# Text Format

A utility packge for writing text files, for cases where you want to avoid fmt.FprintX, offfers some speed advantage.

Simple benchmark provided presents a 29% runtime improvement.

```
goos: windows
goarch: amd64
pkg: github.com/EliCDavis/polyform/formats/txt
cpu: 13th Gen Intel(R) Core(TM) i7-13800H
BenchmarkFormat_Fprintf-20       	 3094824	       484.2 ns/op	     114 B/op	       0 allocs/op
BenchmarkFormat_TextWriter-20    	 3141602	       342.4 ns/op	     189 B/op	       8 allocs/op
```