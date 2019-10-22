package main

import (
	"github.com/linkpoolio/bridges"
	"net/http"
)

// CryptoCompare is the most basic Bridge implementation, as it only calls the api:
// https://min-api.cryptocompare.com/data/price?fsym=ETH&tsyms=USD,JPY,EUR
type CryptoCompare struct{}

// Run is the bridge.Bridge Run implementation that returns the price response
func (cc *CryptoCompare) Run(h *bridges.Helper) (interface{}, error) {
	r := make(map[string]interface{})
	err := h.HTTPCall(
		http.MethodGet,
		"https://min-api.cryptocompare.com/data/price?fsym=ETH&tsyms=USD,JPY,EUR",
		&r,
	)
	return r, err
}

// Opts is the bridge.Bridge implementation
func (cc *CryptoCompare) Opts() *bridges.Opts {
	return &bridges.Opts{
		Name:   "CryptoCompare",
		Lambda: true,
	}
}

func main() {
	bridges.NewServer(&CryptoCompare{}).Start(8080)
}
