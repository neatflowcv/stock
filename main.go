package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

const (
	defaultAlpacaURL  = "https://paper-api.alpaca.markets"
	yahooQuoteURL     = "https://query1.finance.yahoo.com/v7/finance/quote"
	defaultLimit      = 30
	httpTimeout       = 15 * time.Second
	yahooBatchSize    = 200
	responseBodyLimit = 1024
)

var (
	errLimitMustBePositive     = errors.New("limit must be > 0")
	errMissingAlpacaAPIKeys    = errors.New("set APCA_API_KEY_ID and APCA_API_SECRET_KEY env vars")
	errNoSymbolsAfterFiltering = errors.New("no symbols left after filtering")
	errUnknownFormat           = errors.New("unknown format")
	errAlpacaAPIStatus         = errors.New("alpaca API status")
	errYahooAPIStatus          = errors.New("yahoo API status")
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
	Symbol   string  `json:"symbol"`
	Price    float64 `json:"price"`
	Currency string  `json:"currency"`
	State    string  `json:"state"`
}

type App struct {
	Client        *http.Client
	AlpacaBaseURL string
	YahooQuoteURL string
}

func main() {
	err := run()
	if err != nil {
		exitErr(err)
	}
}

func run() error {
	limit, symbolsFilter, outputFormat, alpacaURL := parseFlags()

	if limit <= 0 {
		return errLimitMustBePositive
	}

	keyID, secretKey, credentialErr := readCredentials()
	if credentialErr != nil {
		return credentialErr
	}

	app := newApp(alpacaURL)
	ctx := context.Background()

	symbols, err := app.FetchFractionableSymbols(ctx, keyID, secretKey)
	if err != nil {
		return err
	}

	symbols = filterSymbols(symbols, parseSymbolFilter(symbolsFilter))
	if len(symbols) == 0 {
		return errNoSymbolsAfterFiltering
	}

	if len(symbols) > limit {
		symbols = symbols[:limit]
	}

	prices, err := app.FetchYahooPrices(ctx, symbols)
	if err != nil {
		return err
	}

	rows := buildRows(symbols, prices)

	return writeOutput(rows, outputFormat)
}

func parseFlags() (int, string, string, string) {
	limit := flag.Int("limit", defaultLimit, "Maximum number of symbols to print")
	symbols := flag.String("symbols", "", "Comma-separated symbols to include (optional)")
	format := flag.String("format", "table", "Output format: table or json")
	alpacaURL := flag.String("alpaca-url", defaultAlpacaURL, "Alpaca API base URL")

	flag.Parse()

	return *limit, *symbols, *format, *alpacaURL
}

func readCredentials() (string, string, error) {
	keyID := strings.TrimSpace(os.Getenv("APCA_API_KEY_ID"))

	secretKey := strings.TrimSpace(os.Getenv("APCA_API_SECRET_KEY"))

	if keyID == "" || secretKey == "" {
		return "", "", errMissingAlpacaAPIKeys
	}

	return keyID, secretKey, nil
}

func newApp(alpacaURL string) *App {
	return &App{
		Client: &http.Client{
			Transport:     nil,
			CheckRedirect: nil,
			Jar:           nil,
			Timeout:       httpTimeout,
		},
		AlpacaBaseURL: strings.TrimRight(alpacaURL, "/"),
		YahooQuoteURL: yahooQuoteURL,
	}
}

func buildRows(symbols []string, prices map[string]PriceRow) []PriceRow {
	rows := make([]PriceRow, 0, len(symbols))

	for _, symbol := range symbols {
		price, exists := prices[symbol]
		if exists {
			rows = append(rows, price)
		}
	}

	return rows
}

func writeOutput(rows []PriceRow, outputFormat string) error {
	switch strings.ToLower(strings.TrimSpace(outputFormat)) {
	case "json":
		bytes, err := json.MarshalIndent(rows, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal output JSON: %w", err)
		}

		bytes = append(bytes, '\n')

		_, err = os.Stdout.Write(bytes)
		if err != nil {
			return fmt.Errorf("write JSON output: %w", err)
		}
	case "table":
		printTable(rows)
	default:
		return fmt.Errorf("%w: %s", errUnknownFormat, outputFormat)
	}

	return nil
}

func (app *App) FetchFractionableSymbols(ctx context.Context, keyID, secretKey string) ([]string, error) {
	assetsURL, err := url.Parse(app.AlpacaBaseURL + "/v2/assets")
	if err != nil {
		return nil, fmt.Errorf("parse alpaca URL: %w", err)
	}

	query := assetsURL.Query()
	query.Set("status", "active")
	query.Set("asset_class", "us_equity")
	assetsURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, assetsURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("build alpaca request: %w", err)
	}

	req.Header.Set("Apca-Api-Key-Id", keyID)
	req.Header.Set("Apca-Api-Secret-Key", secretKey)

	//nolint:gosec // URL is built from configured base URL and encoded query params.
	resp, err := app.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("perform alpaca request: %w", err)
	}

	assets, err := parseAlpacaAssetsResponse(resp)
	if err != nil {
		return nil, err
	}

	symbols := make([]string, 0, len(assets))

	for _, asset := range assets {
		if asset.Status == "active" && asset.Tradable && asset.Fractionable {
			symbols = append(symbols, strings.ToUpper(strings.TrimSpace(asset.Symbol)))
		}
	}

	sort.Strings(symbols)

	return symbols, nil
}

func parseAlpacaAssetsResponse(resp *http.Response) ([]alpacaAsset, error) {
	if resp.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, responseBodyLimit))

		closeErr := resp.Body.Close()

		if readErr != nil {
			return nil, joinErr(
				fmt.Errorf("read alpaca error response: %w", readErr),
				wrapCloseErr(closeErr, "alpaca response body"),
			)
		}

		return nil, joinErr(
			fmt.Errorf("%w %d: %s", errAlpacaAPIStatus, resp.StatusCode, strings.TrimSpace(string(body))),
			wrapCloseErr(closeErr, "alpaca response body"),
		)
	}

	var assets []alpacaAsset

	decodeErr := json.NewDecoder(resp.Body).Decode(&assets)

	closeErr := resp.Body.Close()

	if decodeErr != nil {
		return nil, joinErr(
			fmt.Errorf("decode alpaca assets: %w", decodeErr),
			wrapCloseErr(closeErr, "alpaca response body"),
		)
	}

	if closeErr != nil {
		return nil, wrapCloseErr(closeErr, "alpaca response body")
	}

	return assets, nil
}

func (app *App) FetchYahooPrices(ctx context.Context, symbols []string) (map[string]PriceRow, error) {
	rowsBySymbol := make(map[string]PriceRow, len(symbols))

	for startIndex := 0; startIndex < len(symbols); startIndex += yahooBatchSize {
		endIndex := min(startIndex+yahooBatchSize, len(symbols))

		batchRows, err := app.fetchYahooPriceBatch(ctx, symbols[startIndex:endIndex])
		if err != nil {
			return nil, err
		}

		maps.Copy(rowsBySymbol, batchRows)
	}

	return rowsBySymbol, nil
}

func (app *App) fetchYahooPriceBatch(ctx context.Context, symbols []string) (map[string]PriceRow, error) {
	quoteURL, err := url.Parse(app.YahooQuoteURL)
	if err != nil {
		return nil, fmt.Errorf("parse yahoo quote URL: %w", err)
	}

	query := quoteURL.Query()
	query.Set("symbols", strings.Join(symbols, ","))
	quoteURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, quoteURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("build yahoo quote request: %w", err)
	}

	//nolint:gosec // URL is built from configured base URL and encoded query params.
	resp, err := app.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("perform yahoo quote request: %w", err)
	}

	response, err := parseYahooQuoteResponse(resp)
	if err != nil {
		return nil, err
	}

	rowsBySymbol := make(map[string]PriceRow, len(response.QuoteResponse.Result))

	for _, result := range response.QuoteResponse.Result {
		symbol := strings.ToUpper(strings.TrimSpace(result.Symbol))
		rowsBySymbol[symbol] = PriceRow{
			Symbol:   symbol,
			Price:    result.RegularMarketPrice,
			Currency: result.Currency,
			State:    result.MarketState,
		}
	}

	return rowsBySymbol, nil
}

func parseYahooQuoteResponse(resp *http.Response) (yahooResponse, error) {
	if resp.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, responseBodyLimit))

		closeErr := resp.Body.Close()

		if readErr != nil {
			return yahooResponse{}, joinErr(
				fmt.Errorf("read yahoo error response: %w", readErr),
				wrapCloseErr(closeErr, "yahoo response body"),
			)
		}

		return yahooResponse{}, joinErr(
			fmt.Errorf("%w %d: %s", errYahooAPIStatus, resp.StatusCode, strings.TrimSpace(string(body))),
			wrapCloseErr(closeErr, "yahoo response body"),
		)
	}

	var response yahooResponse

	decodeErr := json.NewDecoder(resp.Body).Decode(&response)

	closeErr := resp.Body.Close()

	if decodeErr != nil {
		return yahooResponse{}, joinErr(
			fmt.Errorf("decode yahoo response: %w", decodeErr),
			wrapCloseErr(closeErr, "yahoo response body"),
		)
	}

	if closeErr != nil {
		return yahooResponse{}, wrapCloseErr(closeErr, "yahoo response body")
	}

	return response, nil
}

func wrapCloseErr(closeErr error, resource string) error {
	if closeErr == nil {
		return nil
	}

	return fmt.Errorf("close %s: %w", resource, closeErr)
}

func joinErr(primaryErr, secondaryErr error) error {
	if primaryErr == nil {
		return secondaryErr
	}

	if secondaryErr == nil {
		return primaryErr
	}

	return errors.Join(primaryErr, secondaryErr)
}

func parseSymbolFilter(rawFilter string) map[string]struct{} {
	rawFilter = strings.TrimSpace(rawFilter)
	if rawFilter == "" {
		return nil
	}

	filteredSymbols := make(map[string]struct{})

	for part := range strings.SplitSeq(rawFilter, ",") {
		symbol := strings.ToUpper(strings.TrimSpace(part))
		if symbol != "" {
			filteredSymbols[symbol] = struct{}{}
		}
	}

	if len(filteredSymbols) == 0 {
		return nil
	}

	return filteredSymbols
}

func filterSymbols(symbols []string, filters map[string]struct{}) []string {
	if len(filters) == 0 {
		return symbols
	}

	filtered := make([]string, 0, len(symbols))

	for _, symbol := range symbols {
		if _, exists := filters[symbol]; exists {
			filtered = append(filtered, symbol)
		}
	}

	return filtered
}

func printTable(rows []PriceRow) {
	_, _ = fmt.Fprintf(os.Stdout, "%-8s %-12s %-10s %-12s\n", "SYMBOL", "PRICE", "CURRENCY", "STATE")
	for _, row := range rows {
		_, _ = fmt.Fprintf(os.Stdout, "%-8s %-12.2f %-10s %-12s\n", row.Symbol, row.Price, row.Currency, row.State)
	}
}

func exitErr(err error) {
	_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)

	os.Exit(1)
}
