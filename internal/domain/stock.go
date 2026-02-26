package domain

import (
	"fmt"
	"slices"
	"strings"
	"time"
)

type Stock struct {
	symbol     string
	dailyOHLCs []*DailyOHLC
}

func NewStock(symbol string) (*Stock, error) {
	s := strings.ToUpper(strings.TrimSpace(symbol))
	if s == "" {
		return nil, fmt.Errorf("symbol is required")
	}
	return &Stock{symbol: s}, nil
}

func (s *Stock) Symbol() string { return s.symbol }

func (s *Stock) AddDailyOHLC(date time.Time, high, low, close float64) error {
	ohlc, err := NewDailyOHLC(s.symbol, date, high, low, close)
	if err != nil {
		return err
	}
	s.dailyOHLCs = append(s.dailyOHLCs, ohlc)
	return nil
}

func (s *Stock) DailyOHLCs() []*DailyOHLC {
	return slices.Clone(s.dailyOHLCs)
}
