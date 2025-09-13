/*
Copyright © 2025 bogdan alpha.re9@gmail.com
*/
package cmd

import (
	"bufio"
	"crypto/x509"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"sync"

	"github.com/olekukonko/tablewriter"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ping_endpoints",
	Short: "A utility for checking the HTTP status of a list of URLs and displaying the results as a table.",
	Run:   func(cmd *cobra.Command, args []string) { RunApp(cmd, args) },
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}

}

func init() {
	rootCmd.PersistentFlags().StringP("agent", "a", "", "Customs User-Agent for network requests")
	rootCmd.PersistentFlags().StringP("file", "f", "", "Path to file with urls")
}

var data = [][]string{
	{"Site", "Status code", "Description"},
}

type Row struct {
	Domen  string
	Kind   string
	Detail string
}

var r, _ = regexp.Compile(`(https?://)([[:alpha:].]+)([[:graph:]]+)`)

func RunApp(cmd *cobra.Command, args []string) {
	client := &http.Client{}
	userAgent, _ := cmd.Flags().GetString("agent")
	filePath, _ := cmd.Flags().GetString("file")
	var urls []string
	if filePath != "" {
		urls = urlFromFile(filePath)
	} else {
		urls = urlFromStdin(cmd)
	}

	bar := progressbar.Default(int64(len(urls)), "Analyzing given sites")

	results := make(chan Row, 64)
	var wg sync.WaitGroup

	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			var resp *http.Response
			var err error
			if userAgent != "" {
				resp, err = fetch(url, client, userAgent)
			} else {
				resp, err = fetch(url, client, "")
			}

			kind, detail := classifyHTTP(resp, err)
			results <- Row{parseDomen(url), kind, detail}
			bar.Add(1)
		}(url)

	}
	wg.Wait()
	close(results)
	fmt.Printf("\n\n")

	for r := range results {
		data = append(data, []string{r.Domen, r.Kind, r.Detail})
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.Header(data[0])
	table.Bulk(data[1:])
	table.Render()
}

func urlFromFile(filePath string) []string {
	urls := make([]string, 0)
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Failed to read file %v: ", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		urls = append(urls, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Scanner error: %v", err)
	}

	return urls
}

func urlFromStdin(input *cobra.Command) []string {
	urls := make([]string, 0)

	fi, err := os.Stdin.Stat()
	if err != nil {
		log.Printf("Error reading stdin %v: ", err)
		os.Exit(1)
	}
	if (fi.Mode() & os.ModeCharDevice) != 0 {
		log.Println("Stdin stream is empty.")
		os.Exit(0)
	}

	scanner := bufio.NewScanner(input.InOrStdin())
	for scanner.Scan() {
		urls = append(urls, scanner.Text())
	}
	return urls
}

func fetch(url string, client *http.Client, customAgent string) (*http.Response, error) {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	if customAgent != "" {
		req.Header.Set("User-Agent", customAgent)
	} else {
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:128.0) Gecko/20100101 Firefox/128.0")
	}

	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return resp, nil
}

func classifyHTTP(resp *http.Response, err error) (string, string) {
	var kind, detail string
	if resp != nil {
		kind, detail = classifyHTTPStatusCode(resp)
	} else {
		kind, detail = classifyHTTPError(err)
	}
	return kind, detail
}

func parseDomen(url string) string {
	pieces := r.FindStringSubmatch(url)
	return pieces[2]
}

func classifyHTTPError(err error) (kind, detail string) {
	if err == nil {
		return "", ""
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return "timeout", "network timeout (dial/tls/headers/total)"
	}

	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return "dns_error", dnsErr.Error()
	}

	var uaErr x509.UnknownAuthorityError
	if errors.As(err, &uaErr) {
		return "tls_cert", "x509: unknown authority"
	}

	var hnErr x509.HostnameError
	if errors.As(err, &hnErr) {
		return "tls_cert", "x509: hostname mismatch"
	}

	var invErr x509.CertificateInvalidError
	if errors.As(err, &invErr) {
		return "tls_cert", "x509: invalid certificate"
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return "net_op_error", opErr.Err.Error()
	}

	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return "url_error", urlErr.Op + " " + urlErr.URL + ": " + urlErr.Err.Error()
	}

	return "unknown", err.Error()
}

func classifyHTTPStatusCode(resp *http.Response) (code, detail string) {
	if resp.StatusCode == http.StatusOK {
		return strconv.Itoa(resp.StatusCode), "OK"
	}

	// Status code 301
	if resp.StatusCode == http.StatusMovedPermanently {
		return strconv.Itoa(resp.StatusCode), "Moved permanently"
	}

	// Status code 403
	if resp.StatusCode == http.StatusForbidden {
		return strconv.Itoa(resp.StatusCode), "Forbidden"
	}

	// Status code 404
	if resp.StatusCode == http.StatusNotFound {
		return strconv.Itoa(resp.StatusCode), "Not Found"
	}

	// Status code 429
	if resp.StatusCode == http.StatusTooManyRequests {
		return strconv.Itoa(resp.StatusCode), "Too many requests"
	}

	// Status code 500
	if resp.StatusCode == http.StatusInternalServerError {
		return strconv.Itoa(resp.StatusCode), "Internal server error"
	}

	// Status code 502
	if resp.StatusCode == http.StatusBadGateway {
		return strconv.Itoa(resp.StatusCode), "Bad gateway"
	}

	return "unknown code", "unknown"
}
