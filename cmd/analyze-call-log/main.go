package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

type entry struct {
	Time          string          `json:"time"`
	Tool          string          `json:"tool"`
	DurationMs    int64           `json:"durationMs"`
	Arguments     json.RawMessage `json:"arguments"`
	IsError       bool            `json:"isError"`
	Error         string          `json:"error"`
	ResultBytes   int             `json:"resultBytes"`
	ResultPreview string          `json:"resultPreview"`
}

func main() {
	path := os.Args[1]
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var entries []entry
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		var e entry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			fmt.Println("PARSE ERROR:", err, line)
			continue
		}
		entries = append(entries, e)
	}

	fmt.Printf("=== Total calls: %d ===\n\n", len(entries))

	// Count by tool
	counts := map[string]int{}
	totalDuration := map[string]int64{}
	for _, e := range entries {
		counts[e.Tool]++
		totalDuration[e.Tool] += e.DurationMs
	}
	type toolCount struct {
		Tool  string
		Count int
		TotalMs int64
	}
	var tc []toolCount
	for t, c := range counts {
		tc = append(tc, toolCount{t, c, totalDuration[t]})
	}
	sort.Slice(tc, func(i, j int) bool { return tc[i].Count > tc[j].Count })
	fmt.Println("=== Calls by tool ===")
	for _, x := range tc {
		fmt.Printf("%4d  %8dms total  %6.1fms avg  %s\n", x.Count, x.TotalMs, float64(x.TotalMs)/float64(x.Count), x.Tool)
	}

	// Errors
	fmt.Println("\n=== Errors ===")
	errCount := 0
	for i, e := range entries {
		if e.IsError || e.Error != "" {
			errCount++
			fmt.Printf("[%d] tool=%s args=%s error=%s preview=%s\n", i, e.Tool, string(e.Arguments), e.Error, e.ResultPreview)
		}
	}
	fmt.Printf("Total errors: %d\n", errCount)

	// search_node_types: queries, regex usage, duplicates
	fmt.Println("\n=== search_node_types queries ===")
	type searchArgs struct {
		Query string `json:"query"`
		Regex bool   `json:"regex"`
		PathPrefix string `json:"pathPrefix"`
	}
	seen := map[string]int{}
	regexCount := 0
	searchTotal := 0
	for _, e := range entries {
		if e.Tool != "search_node_types" {
			continue
		}
		searchTotal++
		var sa searchArgs
		json.Unmarshal(e.Arguments, &sa)
		if sa.Regex {
			regexCount++
		}
		key := sa.Query + "|" + sa.PathPrefix
		seen[key]++
		fmt.Printf("query=%-30q regex=%-5v pathPrefix=%-25q\n", sa.Query, sa.Regex, sa.PathPrefix)
	}
	fmt.Printf("\nTotal search_node_types calls: %d, using regex: %d\n", searchTotal, regexCount)
	fmt.Println("Duplicate (query+pathPrefix) searches:")
	for k, c := range seen {
		if c > 1 {
			fmt.Printf("  %dx: %s\n", c, k)
		}
	}

	// create_node: types
	fmt.Println("\n=== create_node by type ===")
	type createArgs struct {
		Type  string `json:"type"`
		Scope string `json:"scope"`
	}
	typeCounts := map[string]int{}
	createTotal := 0
	for _, e := range entries {
		if e.Tool != "create_node" {
			continue
		}
		createTotal++
		var ca createArgs
		json.Unmarshal(e.Arguments, &ca)
		typeCounts[ca.Type]++
	}
	type typeCount struct {
		Type  string
		Count int
	}
	var tyc []typeCount
	for t, c := range typeCounts {
		tyc = append(tyc, typeCount{t, c})
	}
	sort.Slice(tyc, func(i, j int) bool { return tyc[i].Count > tyc[j].Count })
	for _, x := range tyc {
		fmt.Printf("%4d  %s\n", x.Count, x.Type)
	}
	fmt.Printf("Total create_node calls: %d\n", createTotal)

	// get_node_types batching check
	fmt.Println("\n=== get_node_types batch sizes ===")
	type getArgs struct {
		Types []string `json:"types"`
	}
	singleCount, batchCount := 0, 0
	for _, e := range entries {
		if e.Tool != "get_node_types" {
			continue
		}
		var ga getArgs
		json.Unmarshal(e.Arguments, &ga)
		if len(ga.Types) <= 1 {
			singleCount++
		} else {
			batchCount++
		}
	}
	fmt.Printf("single-type calls: %d, batched (2+) calls: %d\n", singleCount, batchCount)

	// Largest results
	fmt.Println("\n=== Top 10 largest results ===")
	sorted := append([]entry{}, entries...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].ResultBytes > sorted[j].ResultBytes })
	for i := 0; i < 10 && i < len(sorted); i++ {
		fmt.Printf("%8d bytes  %s\n", sorted[i].ResultBytes, sorted[i].Tool)
	}
}
