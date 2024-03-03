package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

type headerSlice []string

type Response struct {
	URL     string
	Status  string
	Context string
}

// regex for extract html tags
const combinedRegex = `<input[^>]*\sname\s*=\s*["']([^"']*)["']|<a\s[^>]*\bhref\s*=\s*["']([^"']*)["'][^>]*>|<form[^>]*\sname\s*=\s*["']([^"']*)["']|<map[^>]*\sname\s*=\s*["']([^"']*)["']|<fieldset[^>]*\sname\s*=\s*["']([^"']*)["']|<output[^>]*\sname\s*=\s*["']([^"']*)["']|<iframe[^>]*\sname\s*=\s*["']([^"']*)["']|<input[^>]*\sid\s*=\s*["']([^"']*)["']|["']([^"']+?)["']\s*:\s*|<object[^>]*\sname\s*=\s*["']([^"']*)["']|<param[^>]*\sname\s*=\s*["']([^"']*)["']|<textarea[^>]*\sname\s*=\s*["']([^"']*)["']|<select[^>]*\sname\s*=\s*["']([^"']*)["']`

// filter regex for filtering output
const filterRegex = `(:|\"|\s|\n|-|\.|\@|\+|\$|\#|\'|\/)`

// html encode replacing + charchters needs to replace with nothing
var htmlEncodedFilters = []string{"&lt;", "lt;", "\\", "&gt;", "gt", "amp;", "?", "%", ";", "&amp;", "&quot;", "quot;", "apos;", "&apos;", "&nbsp;", "nbsp;"}

// methos for header argument
func (hs *headerSlice) String() string {
	return strings.Join(*hs, ", ")
}

func (hs *headerSlice) Set(value string) error {
	*hs = append(*hs, value)
	return nil
}

// http prober
func paramProbe(u string, header []string, delay, timeout int, ch chan<- Response, wg *sync.WaitGroup) {
	defer wg.Done()
	// timeout
	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
	req, _ := http.NewRequest("GET", u, nil)
	// header
	if len(header) != 0 {
		for _, val := range header {
			sp := strings.Split(val, ":")
			if len(sp) == 2 {
				header_key := strings.TrimSpace(sp[0])
				header_val := strings.TrimSpace(sp[1])
				req.Header.Set(header_key, header_val)
			} else {
				return
			}
		}
	}
	// set delay
	if delay > 1 {
		time.Sleep(time.Second * time.Duration(delay))
	}
	// send request
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	// find regex patterns
	findPattern := regexp.MustCompile(combinedRegex)
	// extract keys
	matches := findPattern.FindAllStringSubmatch(string(body), -1)
	for _, match := range matches {
		for _, l := range match[1:] {
			if l != "" {
				// extract query paramater from anchor tag
				queryParams := extractQueryParams(l)
				// filter
				regexFilter, err := regexp.MatchString(filterRegex, l)
				if err != nil {
					return
				}
				if len(queryParams) > 0 {
					for _, m := range queryParams {
						regexFilter, err := regexp.MatchString(filterRegex, m)
						if err != nil {
							return
						}
						if !regexFilter {
							result := removePrefixes(m, htmlEncodedFilters)
							ch <- Response{URL: u, Status: resp.Status, Context: result}
						}
					}
				}
				if !regexFilter {
					result := removePrefixes(l, htmlEncodedFilters)
					ch <- Response{URL: u, Status: resp.Status, Context: result}
				}

			}
		}
	}

}

func removePrefixes(i string, prefixes []string) string {
	result := i
	for _, prefix := range prefixes {
		result = strings.ReplaceAll(result, prefix, "")
	}
	return result
}

func extractQueryParams(urlString string) []string {
	var params []string
	re := regexp.MustCompile(`[?&]([^=&]+)=([^&]*)`)
	matches := re.FindAllStringSubmatch(urlString, -1)

	for _, match := range matches {
		key := match[1]
		params = append(params, key)
	}
	return params
}

func readLines(fileName string) ([]string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

func removeDuplicates(lines []string) []string {
	uniqueLines := make(map[string]struct{})
	var result []string

	for _, line := range lines {
		// Add line to uniqueLines map (sets automatically eliminate duplicates)
		uniqueLines[line] = struct{}{}
	}

	// Convert uniqueLines map back to a slice
	for line := range uniqueLines {
		result = append(result, line)
	}

	return result
}

func writeLines(fileName string, lines []string) error {
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}

	return writer.Flush()
}

func main() {
	var wg sync.WaitGroup

	var (
		host    string
		hosts   string
		timeout int
		delay   int
		path    string
		header  headerSlice
	)
	responseCh := make(chan Response)
	flag.StringVar(&host, "host", "", "single host to process")
	flag.StringVar(&hosts, "hosts", "", "path of hosts file to process")
	flag.StringVar(&path, "output", "", "path to save output")
	flag.Var(&header, "header", "custom header/cookie to include in requests like that: Cookie:value, Origin:value, Host:value")
	flag.IntVar(&timeout, "timeout", 10, "timeout in seconds")
	flag.IntVar(&delay, "delay", -1, "durations between each HTTP requests")

	flag.Parse()

	if host == "" && hosts == "" {
		flag.Usage()
	}

	if host != "" && hosts == "" {
		wg.Add(1)
		go paramProbe(host, header, delay, timeout, responseCh, &wg)
	}

	if hosts != "" && host == "" {
		file, err := os.Open(hosts)

		if err != nil {
			fmt.Println("File You Proveded is not found: ", err)
			os.Exit(0)
		}

		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			url := scanner.Text()
			wg.Add(1)
			go paramProbe(url, header, delay, timeout, responseCh, &wg)
		}
		defer file.Close()
	}

	go func() {
		wg.Wait()
		close(responseCh)
	}()

	if path != "" {
		outputFile, err := os.Create(path)

		if err != nil {
			log.Fatal(err)
		}

		defer outputFile.Close()

		for res := range responseCh {
			_, err := outputFile.WriteString(res.Context + "\n")
			if err != nil {
				log.Fatal(err)
			}
		}
		// remove duplicate lines
		lines, err := readLines(path)

		if err != nil {
			return
		}

		uniqueLines := removeDuplicates(lines)

		err = writeLines(path, uniqueLines)

		if err != nil {
			return
		}
	}

	if path == "" {
		fmt.Println("Please use -output PATH to save output")
		flag.Usage()
	}
}
