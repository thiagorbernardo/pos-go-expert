package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"os"
	"sort"
	"sync"
	"time"
)

type testConfig struct {
	targetURL       string
	totalRequests   int
	concurrency     int
	requestTimeout  time.Duration
	keepAliveIdle   time.Duration
	maxIdleConns    int
	maxConnsPerHost int
}

type testStats struct {
	mu           sync.Mutex
	total        int
	success200   int
	statusCounts map[int]int
}

func (s *testStats) record(statusCode int) {
	s.mu.Lock()
	s.total++
	if statusCode == http.StatusOK {
		s.success200++
	}
	s.statusCounts[statusCode] = s.statusCounts[statusCode] + 1
	s.mu.Unlock()
}

func parseFlags() (testConfig, error) {
	var (
		urlFlag         string
		requestsFlag    int
		concurrencyFlag int
		timeoutFlag     time.Duration
	)

	flag.StringVar(&urlFlag, "url", "", "URL do serviço a ser testado")
	flag.IntVar(&requestsFlag, "requests", 0, "Número total de requests")
	flag.IntVar(&concurrencyFlag, "concurrency", 1, "Número de chamadas simultâneas")
	flag.DurationVar(&timeoutFlag, "timeout", 10*time.Second, "Timeout por request (ex: 5s, 1m)")
	flag.Parse()

	if urlFlag == "" {
		return testConfig{}, fmt.Errorf("parâmetro --url é obrigatório")
	}
	if _, err := neturl.ParseRequestURI(urlFlag); err != nil {
		return testConfig{}, fmt.Errorf("--url inválida: %v", err)
	}
	if requestsFlag <= 0 {
		return testConfig{}, fmt.Errorf("--requests deve ser > 0")
	}
	if concurrencyFlag <= 0 {
		return testConfig{}, fmt.Errorf("--concurrency deve ser > 0")
	}
	if concurrencyFlag > requestsFlag {
		concurrencyFlag = requestsFlag
	}

	cfg := testConfig{
		targetURL:       urlFlag,
		totalRequests:   requestsFlag,
		concurrency:     concurrencyFlag,
		requestTimeout:  timeoutFlag,
		keepAliveIdle:   30 * time.Second,
		maxIdleConns:    2048,
		maxConnsPerHost: 0, // 0 means no limit; tuned by concurrency via client.
	}
	return cfg, nil
}

func buildHTTPClient(cfg testConfig) *http.Client {
	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DisableCompression:    false,
		MaxIdleConns:          cfg.maxIdleConns,
		MaxIdleConnsPerHost:   cfg.maxIdleConns,
		MaxConnsPerHost:       cfg.maxConnsPerHost,
		IdleConnTimeout:       cfg.keepAliveIdle,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	return &http.Client{
		Transport: transport,
		Timeout:   cfg.requestTimeout,
	}
}

func runLoadTest(cfg testConfig) (time.Duration, *testStats) {
	client := buildHTTPClient(cfg)
	defer client.CloseIdleConnections()

	jobs := make(chan struct{})
	var wg sync.WaitGroup

	stats := &testStats{statusCounts: make(map[int]int)}

	worker := func() {
		defer wg.Done()
		for range jobs {
			resp, err := client.Get(cfg.targetURL)
			if err != nil {
				// Treat network/timeout errors as status code 0
				stats.record(0)
				continue
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			stats.record(resp.StatusCode)
		}
	}

	start := time.Now()
	wg.Add(cfg.concurrency)
	for i := 0; i < cfg.concurrency; i++ {
		go worker()
	}

	for i := 0; i < cfg.totalRequests; i++ {
		jobs <- struct{}{}
	}
	close(jobs)
	wg.Wait()
	elapsed := time.Since(start)

	return elapsed, stats
}

func printReport(elapsed time.Duration, stats *testStats) {
	fmt.Println("==== Relatório de Teste de Carga ====")
	fmt.Printf("Tempo total: %s\n", elapsed)
	fmt.Printf("Total de requests: %d\n", stats.total)
	fmt.Printf("HTTP 200: %d\n", stats.success200)

	fmt.Println("Distribuição de códigos de status:")
	// Sort keys for stable output
	keys := make([]int, 0, len(stats.statusCounts))
	for k := range stats.statusCounts {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, code := range keys {
		label := fmt.Sprintf("%d", code)
		if code == 0 {
			label = "erro (timeout/conexão)"
		}
		fmt.Printf("  %s: %d\n", label, stats.statusCounts[code])
	}
}

func main() {
	cfg, err := parseFlags()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Erro:", err)
		flag.Usage()
		os.Exit(2)
	}

	elapsed, stats := runLoadTest(cfg)
	printReport(elapsed, stats)
}
