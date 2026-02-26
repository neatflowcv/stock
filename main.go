package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

const (
	defaultAlpacaURL = "https://paper-api.alpaca.markets"
	yahooQuoteURL    = "https://query1.finance.yahoo.com/v7/finance/quote"
)

type alpacaAsset struct {
	Symbol       string `json:"symbol"`
	Status       string `json:"status"`
	Tradable     bool   `json:"tradable"`
	Fractionable bool   `json:"fractionable"`
}

type quoteResult struct {
	Symbol             string  `json:"symbol"`
	RegularMarketPrice float64 `json:"regularMarketPrice"`
	Currency           string  `json:"currency"`
	MarketState        string  `json:"marketState"`
}

type yahooResponse struct {
	QuoteResponse struct {
		Result []quoteResult `json:"result"`
	} `json:"quoteResponse"`
}

type PriceRow struct {
	Symbol   string
	Price    float64
	Currency string
	State    string
}

type App struct {
	Client        *http.Client
	AlpacaBaseURL string
	YahooQuoteURL string
}

func main() {
	var (
		limit       = flag.Int("limit", 30, "Maximum number of symbols to print")
		symbolsFlag = flag.String("symbols", "", "Comma-separated symbols to include (optional)")
		format      = flag.String("format", "table", "Output format: table or json")
		alpacaURL   = flag.String("alpaca-url", defaultAlpacaURL, "Alpaca API base URL")
	)
	flag.Parse()

	if *limit <= 0 {
		exitErr(errors.New("limit must be > 0"))
	}

	key := strings.TrimSpace(os.Getenv("APCA_API_KEY_ID"))
	secret := strings.TrimSpace(os.Getenv("APCA_API_SECRET_KEY"))
	if key == "" || secret == "" {
		exitErr(errors.New("set APCA_API_KEY_ID and APCA_API_SECRET_KEY env vars"))
	}

	app := &App{
		Client:        &http.Client{Timeout: 15 * time.Second},
		AlpacaBaseURL: strings.TrimRight(*alpacaURL, "/"),
		YahooQuoteURL: yahooQuoteURL,
	}

	ctx := context.Background()

	symbols, err := app.FetchFractionableSymbols(ctx, key, secret)
	if err != nil {
		exitErr(err)
	}

	filter := parseSymbolFilter(*symbolsFlag)
	symbols = filterSymbols(symbols, filter)
	if len(symbols) == 0 {
		exitErr(errors.New("no symbols left after filtering"))
	}

	if len(symbols) > *limit {
		symbols = symbols[:*limit]
	}

	prices, err := app.FetchYahooPrices(ctx, symbols)
	if err != nil {
		exitErr(err)
	}

	rows := make([]PriceRow, 0, len(symbols))
	for _, s := range symbols {
		if p, ok := prices[s]; ok {
			rows = append(rows, p)
		}
	}

	switch strings.ToLower(strings.TrimSpace(*format)) {
	case "json":
		b, err := json.MarshalIndent(rows, "", "  ")
		if err != nil {
			exitErr(err)
		}
		fmt.Println(string(b))
	case "table":
		printTable(rows)
	default:
		exitErr(fmt.Errorf("unknown format: %s", *format))
	}
}

func (a *App) FetchFractionableSymbols(ctx context.Context, key, secret string) ([]string, error) {
	u, err := url.Parse(a.AlpacaBaseURL + "/v2/assets")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("status", "active")
	q.Set("asset_class", "us_equity")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("APCA-API-KEY-ID", key)
	req.Header.Set("APCA-API-SECRET-KEY", secret)

	resp, err := a.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("alpaca API status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var assets []alpacaAsset
	if err := json.NewDecoder(resp.Body).Decode(&assets); err != nil {
		return nil, err
	}

	symbols := make([]string, 0, len(assets))
	for _, as := range assets {
		if as.Status == "active" && as.Tradable && as.Fractionable {
			symbols = append(symbols, strings.ToUpper(strings.TrimSpace(as.Symbol)))
		}
	}
	sort.Strings(symbols)
	return symbols, nil
}

func (a *App) FetchYahooPrices(ctx context.Context, symbols []string) (map[string]PriceRow, error) {
	const batch = 200
	out := make(map[string]PriceRow, len(symbols))

	for i := 0; i < len(symbols); i += batch {
		j := i + batch
		if j > len(symbols) {
			j = len(symbols)
		}
		chunk := symbols[i:j]

		u, err := url.Parse(a.YahooQuoteURL)
		if err != nil {
			return nil, err
		}
		q := u.Query()
		q.Set("symbols", strings.Join(chunk, ","))
		u.RawQuery = q.Encode()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, err
		}

		resp, err := a.Client.Do(req)
		if err != nil {
			return nil, err
		}

		func() {
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				b, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
				err = fmt.Errorf("yahoo API status %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
				return
			}

			var yr yahooResponse
			if decErr := json.NewDecoder(resp.Body).Decode(&yr); decErr != nil {
				err = decErr
				return
			}

			for _, r := range yr.QuoteResponse.Result {
				s := strings.ToUpper(strings.TrimSpace(r.Symbol))
				out[s] = PriceRow{
					Symbol:   s,
					Price:    r.RegularMarketPrice,
					Currency: r.Currency,
					State:    r.MarketState,
				}
			}
		}()
		if err != nil {
			return nil, err
		}
	}

	return out, nil
}

func parseSymbolFilter(s string) map[string]struct{} {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	out := make(map[string]struct{})
	parts := strings.Split(s, ",")
	for _, p := range parts {
		p = strings.ToUpper(strings.TrimSpace(p))
		if p != "" {
			out[p] = struct{}{}
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func filterSymbols(symbols []string, filter map[string]struct{}) []string {
	if len(filter) == 0 {
		return symbols
	}
	out := make([]string, 0, len(symbols))
	for _, s := range symbols {
		if _, ok := filter[s]; ok {
			out = append(out, s)
		}
	}
	return out
}

func printTable(rows []PriceRow) {
	fmt.Printf("%-8s %-12s %-10s %-12s\n", "SYMBOL", "PRICE", "CURRENCY", "STATE")
	for _, r := range rows {
		fmt.Printf("%-8s %-12.2f %-10s %-12s\n", r.Symbol, r.Price, r.Currency, r.State)
	}
}

func exitErr(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}
