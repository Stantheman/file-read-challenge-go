package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

func main() {
	start := time.Now()
	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	firstNamePat := regexp.MustCompile(", \\s*([^, ]+)")
	names := make([]string, 0)
	firstNames := make([]string, 0)
	dates := make([]string, 0)
	commonName := ""
	commonCount := 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()

		// get all the names
		split := strings.SplitN(text, "|", 9)
		name := strings.TrimSpace(split[7])
		names = append(names, name)

		// extract first names
		if matches := firstNamePat.FindAllStringSubmatch(name, 1); len(matches) > 0 {
			firstNames = append(firstNames, matches[0][1])
		}

		// extract dates
		chars := strings.TrimSpace(split[4])[:6]
		date := chars[:4] + "-" + chars[4:6]
		dates = append(dates, date)
	}

	// report c2: names at index
	fmt.Printf("Name: %s at index: %v\n", names[0], 0)
	fmt.Printf("Name: %s at index: %v\n", names[432], 432)
	fmt.Printf("Name: %s at index: %v\n", names[43243], 43243)
	fmt.Printf("Name time: %v\n", time.Since(start))

	// report c1: total number of lines
	fmt.Printf("Total file line count: %v\n", len(names))
	fmt.Printf("Line count time: %v\n", time.Since(start))

	// report c3: donation frequency
	dateMap := make(map[string]int)
	for _, date := range dates {
		dateMap[date] += 1
	}
	for k, v := range dateMap {
		fmt.Printf("Donations per month and year: %v and donation count: %v\n", k, v)
	}
	fmt.Printf("Donations time: %v\n", time.Since(start))

	// report c4: most common firstName
	nameMap := make(map[string]int)
	nameCount := 0 // new count
	for _, name := range firstNames {
		nameCount = nameMap[name] + 1
		nameMap[name] = nameCount
		if nameCount > commonCount {
			commonName = name
			commonCount = nameCount
		}
	}
	fmt.Printf("The most common first name is: %s and it occurs: %v times.\n", commonName, commonCount)
	fmt.Printf("Most common name time: %v\n", time.Since(start))
}
