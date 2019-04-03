package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	start := time.Now()
	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	f, err := os.Create("/tmp/pprof.out")
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	commonName := ""
	commonCount := 0
	scanner := bufio.NewScanner(bufio.NewReaderSize(file, 65536))
	nameMap := make(map[string]int)
	dateMap := make(map[int]int)

	namesCounted := false
	namesCount := 0
	fileLineCount := int64(0)

	type entry struct {
		firstName string
		name      string
		date      int
	}

	linesChunkLen := 64 * 1024
	linesChunkPoolAllocated := int64(0)
	linesPool := sync.Pool{New: func() interface{} {
		lines := make([]string, 0, linesChunkLen)
		atomic.AddInt64(&linesChunkPoolAllocated, 1)
		return lines
	}}
	lines := linesPool.Get().([]string)[:0]

	entriesPoolAllocated := int64(0)
	entriesPool := sync.Pool{New: func() interface{} {
		entries := make([]entry, 0, linesChunkLen)
		atomic.AddInt64(&entriesPoolAllocated, 1)
		return entries
	}}
	mutex := &sync.Mutex{}
	wg := sync.WaitGroup{}

	scanner.Scan()
	for {
		lines = append(lines, scanner.Text())
		willScan := scanner.Scan()
		if len(lines) == linesChunkLen || !willScan {
			linesToProcess := lines
			wg.Add(len(linesToProcess))
			go func() {
				atomic.AddInt64(&fileLineCount, int64(len(linesToProcess)))
				entries := entriesPool.Get().([]entry)[:0]
				for _, text := range linesToProcess {
					// get all the names
					entry := entry{}

					// get the name and date, field 7 and 4
					new_name, new_date := "", ""
					for offset, count := 0, 0; ; {
						next_offset := strings.Index(text[offset:], "|")
						if next_offset == -1 {
							fmt.Printf("fuckk\n")
							break
						}
						next_offset += offset
						count++
						if count == 5 {
							new_date = text[offset:next_offset]
						}
						if count == 8 {
							new_name = text[offset:next_offset]
						}
						if count > 8 {
							break
						}
						offset = next_offset + 1
					}
					entry.name = strings.TrimSpace(new_name)

					// extract first names
					if entry.name != "" {
						startOfName := strings.Index(entry.name, ", ") + 2
						if endOfName := strings.IndexByte(entry.name[startOfName:], ' '); endOfName < 0 {
							entry.firstName = entry.name[startOfName:]
						} else {
							entry.firstName = entry.name[startOfName : startOfName+endOfName]
						}
						if cs := strings.IndexByte(entry.firstName, ','); cs > 0 {
							entry.firstName = entry.firstName[:cs]
						}
					}
					// extract dates
					entry.date, _ = strconv.Atoi(new_date[:6])
					entries = append(entries, entry)
				}
				linesPool.Put(linesToProcess)
				mutex.Lock()
				for _, entry := range entries {
					if len(entry.firstName) != 0 {
						nameCount := nameMap[entry.firstName] + 1
						nameMap[entry.firstName] = nameCount
						if nameCount > commonCount {
							commonCount = nameCount
							commonName = entry.firstName
						}
					}
					if namesCounted == false {
						if namesCount == 0 {
							fmt.Printf("Name: %s at index: %v\n", entry.name, 0)
						} else if namesCount == 432 {
							fmt.Printf("Name: %s at index: %v\n", entry.name, 432)
						} else if namesCount == 43243 {
							fmt.Printf("Name: %s at index: %v\n", entry.name, 43243)
							namesCounted = true
						}
						namesCount++
					}
					dateMap[entry.date]++
				}
				mutex.Unlock()
				entriesPool.Put(entries)
				wg.Add(-len(entries))
			}()
			lines = linesPool.Get().([]string)[:0]
		}
		if !willScan {
			break
		}
	}
	wg.Wait()

	// report c2: names at index
	fmt.Printf("Name time: %v\n", time.Since(start))

	// report c1: total number of lines
	fmt.Printf("Total file line count: %v\n", fileLineCount)
	fmt.Printf("Line count time: %v\n", time.Since(start))

	// report c3: donation frequency
	for k, v := range dateMap {
		fmt.Printf("Donations per month and year: %v and donation count: %v\n", k, v)
	}
	fmt.Printf("Donations time: %v\n", time.Since(start))

	// report c4: most common firstName
	fmt.Printf("The most common first name is: %s and it occurs: %v times.\n", commonName, commonCount)
	fmt.Printf("Most common name time: %v\n", time.Since(start))
}