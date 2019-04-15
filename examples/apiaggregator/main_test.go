package main

import (
	"github.com/linkpoolio/bridges/bridge"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAPIAggregator_Run(t *testing.T) {
	aggregationTypes := []string{
		"", "mean", "median", "mode", "nonexistent",
	}
	aa := APIAggregator{}
	for _, at := range aggregationTypes {
		t.Run(at, func(t *testing.T) {
			p := map[string]interface{}{
				"api": []string{
					"https://www.bitstamp.net/api/v2/ticker/btcusd/",
					"https://api.pro.coinbase.com/products/btc-usd/ticker",
				},
				"paths": []string{
					"$.last",
					"$.price",
				},
				"type": at,
			}

			json, err := bridge.ParseInterface(p)
			assert.Nil(t, err)

			h := bridge.NewHelper(json)
			obj, err := aa.Run(h)
			assert.Nil(t, err)

			r, ok := obj.(Result)
			assert.True(t, ok)

			assert.True(t, len(r.AggregateValue) != 0)
			assert.True(t, len(r.APIErrors) == 0)
			assert.Equal(t, r.AggregationType, at)
		})
	}
}

func TestFetch_EmptyParam(t *testing.T) {
	p := map[string]interface{}{}
	aa := APIAggregator{}
	json, err := bridge.ParseInterface(p)
	assert.Nil(t, err)

	h := bridge.NewHelper(json)
	_, err = aa.Run(h)

	assert.Equal(t, err.Error(), "Invalid api and path array")
}

func TestFetch_InvalidArray(t *testing.T) {
	p := map[string]interface{}{
		"api": []string{
			"https://www.bitstamp.net/api/v2/ticker/btcusd/",
			"https://api.pro.coinbase.com/products/btc-usd/ticker",
		},
		"paths": []string{
			"$.last",
		},
		"type": "mode",
	}
	aa := APIAggregator{}
	json, err := bridge.ParseInterface(p)
	assert.Nil(t, err)

	h := bridge.NewHelper(json)
	_, err = aa.Run(h)

	assert.Equal(t, err.Error(), "Invalid api and path array")
}

func TestCryptoCompare_Opts(t *testing.T) {
	cc := APIAggregator{}
	opts := cc.Opts()
	assert.Equal(t, opts.Name, "APIAggregator")
	assert.True(t, opts.Lambda)
}
