package domain

import (
	"errors"
	"slices"
	"strings"
	"time"
)

var errSymbolRequired = errors.New("symbol is required")

type Stock struct {
	symbol     string
	dailyOHLCs []*DailyOHLC
}

func NewStock(symbol string) (*Stock, error) {
	normalizedSymbol := strings.ToUpper(strings.TrimSpace(symbol))
	if normalizedSymbol == "" {
		return nil, errSymbolRequired
	}

	return &Stock{
		symbol:     normalizedSymbol,
		dailyOHLCs: nil,
	}, nil
}

func (s *Stock) Symbol() string { return s.symbol }

func (s *Stock) AddDailyOHLC(date time.Time, open, high, low, closePrice float64) error {
	ohlc, err := NewDailyOHLC(s.symbol, date, open, high, low, closePrice)
	if err != nil {
		return err
	}

	s.dailyOHLCs = append(s.dailyOHLCs, ohlc)

	return nil
}

func (s *Stock) DailyOHLCs() []*DailyOHLC {
	return slices.Clone(s.dailyOHLCs)
}
