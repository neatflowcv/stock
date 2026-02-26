package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	errStockSymbolRequired = errors.New("stock symbol is required")
	errDateRequired        = errors.New("date is required")
	errPriceMustBePositive = errors.New("price must be >= 0")
	errLowGreaterThanHigh  = errors.New("low must be <= high")
)

type DailyOHLC struct {
	stockSymbol string
	date        time.Time
	open        float64
	high        float64
	low         float64
	closePrice  float64
}

func NewDailyOHLC(stockSymbol string, date time.Time, open, high, low, closePrice float64) (*DailyOHLC, error) {
	symbol := strings.ToUpper(strings.TrimSpace(stockSymbol))
	if symbol == "" {
		return nil, errStockSymbolRequired
	}

	if date.IsZero() {
		return nil, errDateRequired
	}

	if open < 0 || high < 0 || low < 0 || closePrice < 0 {
		return nil, errPriceMustBePositive
	}

	if low > high {
		return nil, errLowGreaterThanHigh
	}

	return &DailyOHLC{
		stockSymbol: symbol,
		date:        date,
		open:        open,
		high:        high,
		low:         low,
		closePrice:  closePrice,
	}, nil
}

func (d *DailyOHLC) StockSymbol() string { return d.stockSymbol }

func (d *DailyOHLC) Date() time.Time { return d.date }

func (d *DailyOHLC) Open() float64 { return d.open }

func (d *DailyOHLC) High() float64 { return d.high }

func (d *DailyOHLC) Low() float64 { return d.low }

func (d *DailyOHLC) Close() float64 { return d.closePrice }
