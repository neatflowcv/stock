package domain

import (
	"fmt"
	"strings"
	"time"
)

type DailyOHLC struct {
	stockSymbol string
	date        time.Time
	high        float64
	low         float64
	close       float64
}

func NewDailyOHLC(stockSymbol string, date time.Time, high, low, close float64) (*DailyOHLC, error) {
	symbol := strings.ToUpper(strings.TrimSpace(stockSymbol))
	if symbol == "" {
		return nil, fmt.Errorf("stock symbol is required")
	}
	if date.IsZero() {
		return nil, fmt.Errorf("date is required")
	}
	if high < 0 || low < 0 || close < 0 {
		return nil, fmt.Errorf("price must be >= 0")
	}
	if low > high {
		return nil, fmt.Errorf("low must be <= high")
	}

	return &DailyOHLC{
		stockSymbol: symbol,
		date:        date,
		high:        high,
		low:         low,
		close:       close,
	}, nil
}

func (d *DailyOHLC) StockSymbol() string { return d.stockSymbol }

func (d *DailyOHLC) Date() time.Time { return d.date }

func (d *DailyOHLC) High() float64 { return d.high }

func (d *DailyOHLC) Low() float64 { return d.low }

func (d *DailyOHLC) Close() float64 { return d.close }
