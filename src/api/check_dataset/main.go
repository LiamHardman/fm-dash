package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"

	pb "api/proto"

	"google.golang.org/protobuf/proto"
)

func main() {
	f, err := os.Open("../datasets/6cd07fd9-0e6e-4a6b-acdc-8a77af11d4a5.pb.gz")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		panic(err)
	}
	defer gz.Close()

	data, err := io.ReadAll(gz)
	if err != nil {
		panic(err)
	}

	var ds pb.DatasetData
	if err := proto.Unmarshal(data, &ds); err != nil {
		panic(err)
	}

	// Group by Division -> set of BasedIn values
	divCountries := make(map[string]map[string]int)
	for _, p := range ds.Players {
		if p.Division == "" {
			continue
		}
		if divCountries[p.Division] == nil {
			divCountries[p.Division] = make(map[string]int)
		}
		divCountries[p.Division][p.BasedIn]++
	}

	fmt.Println("Divisions with multiple BasedIn values (ambiguous):")
	for div, countries := range divCountries {
		if len(countries) > 1 {
			fmt.Printf("  %q:\n", div)
			for country, count := range countries {
				fmt.Printf("    BasedIn=%q  players=%d\n", country, count)
			}
		}
	}

	fmt.Println("\nDivisions with single BasedIn (sample of large ones):")
	for div, countries := range divCountries {
		if len(countries) == 1 {
			total := 0
			country := ""
			for c, n := range countries { country = c; total = n }
			if total > 200 {
				fmt.Printf("  %q -> BasedIn=%q players=%d\n", div, country, total)
			}
		}
	}
}
