package render

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func parseStatsBlock(r io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(r)
	inStats := false
	inQuery := false
	lines := make([]string, 0, 1024)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "//-- starting rendering timer..." {
			inStats = true
			continue
		}
		if line == "//-- rendering timer stopped..." {
			return lines, nil
		}
		if inStats {
			// concatenate all lines between postgis_connection::execute_query and postgis_datasource::features
			if strings.Contains(line, "postgis_datasource::features") {
				inQuery = false
			}
			if inQuery {
				lines[len(lines)-1] += " " + line[4:]
			} else {
				lines = append(lines, line)
			}
			if strings.Contains(line, "postgis_connection::execute_query") {
				inQuery = true
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return nil, nil
}

var statsPrefix = `(\d+.\d+ms) \(cpu (?:\d+.\d+ms)\) \s+ \| `
var statsTotal = regexp.MustCompile(statsPrefix + `rendering total for layer: '(.*?)'`)
var statsStyle = regexp.MustCompile(statsPrefix + `rendering style for layer: '(.*?)' and style '(.*?)'`)
var pgQuery = regexp.MustCompile(statsPrefix + `postgis_connection::execute_query (.*)`)

var statsOther = regexp.MustCompile(statsPrefix + `(.*?\n(?:    .*?\n)*)`)

var percentRendered = regexp.MustCompile(`percent rendered: 100% - \d+ rendered for (\d+) queried for  layer '(.*?)' and style '(.*?)'`)

type layerStats struct {
	total         time.Duration
	styles        map[string]int
	queryDuration time.Duration
	query         string
}

func parseStats(lines []string, withSQL bool) error {
	var err error
	var lastQuery string
	var lastQueryDuration time.Duration

	layers := make(map[string]layerStats)
	ordered := []string{}

	for _, line := range lines {
		if match := pgQuery.FindStringSubmatch(line); match != nil {
			lastQueryDuration, err = time.ParseDuration(match[1])
			if err != nil {
				return err
			}
			lastQuery = match[2]
		} else if match := statsStyle.FindStringSubmatch(line); match != nil {
			if l, ok := layers[match[2]]; ok {
				l.styles[match[3]] = 0
				layers[match[2]] = l
			} else {
				layers[match[2]] = layerStats{styles: map[string]int{match[3]: 0}}
				ordered = append(ordered, match[2])
			}
		} else if match := statsTotal.FindStringSubmatch(line); match != nil {
			duration, err := time.ParseDuration(match[1])
			if err != nil {
				return err
			}
			if l, ok := layers[match[2]]; ok {
				l.total = duration
				layers[match[2]] = l
			} else {
				layers[match[2]] = layerStats{total: duration}
				ordered = append(ordered, match[2])
			}
			if lastQuery != "" {
				l := layers[match[2]]
				l.query = lastQuery
				l.queryDuration = lastQueryDuration
				lastQuery = ""
				layers[match[2]] = l
			}
		} else if match := percentRendered.FindStringSubmatch(line); match != nil {
			numFeatures, _ := strconv.ParseInt(match[1], 10, 32)
			if l, ok := layers[match[2]]; ok {
				l.styles[match[3]] = int(numFeatures)
				layers[match[2]] = l
			} else {
				layers[match[2]] = layerStats{styles: map[string]int{match[3]: int(numFeatures)}}
				ordered = append(ordered, match[2])
			}
		}
	}

	var totalQueryDuration time.Duration
	var totalStyleRender time.Duration
	var totalRender time.Duration

	for _, stats := range layers {
		totalRender += stats.total
		if stats.styles != nil {
			totalStyleRender += stats.total
			totalQueryDuration += stats.queryDuration
		}
	}
	fmt.Println(totalRender, totalStyleRender, totalQueryDuration)
	numStyled := 0
	for _, layer := range ordered {
		stats := layers[layer]
		if stats.styles != nil {
			numStyled += 1
			fmt.Printf("%-40s %5.1fms %4.1f%% %5.1fms %4.1f%%", layer, stats.total.Seconds()*1000.0, float64(stats.total)/float64(totalStyleRender)*100, stats.queryDuration.Seconds()*1000, float64(stats.queryDuration)/float64(totalQueryDuration)*100)
			if withSQL {
				fmt.Print(" ", stats.query)
			}
			fmt.Print("\n")
			for style, features := range stats.styles {
				fmt.Printf("\t%-40s %6d\n", style, features)
			}
		}
	}
	fmt.Println(numStyled, len(ordered))
	return nil
}

func main1() {
	var withSQL = flag.Bool("sql", false, "output SQL queries")
	flag.Parse()
	for {
		lines, err := parseStatsBlock(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
		if lines == nil {
			break
		}
		parseStats(lines, *withSQL)
	}
}
