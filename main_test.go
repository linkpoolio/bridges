package main

import (
	"github.com/jleeh/bridges/bridge"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewJSONBridges(t *testing.T) {
	cases := []struct {
		path         string
		name         string
		req          map[string]interface{}
		expectedKeys []string
	}{
		{"json/cryptocompare.json", "CryptoCompare", nil, []string{"EUR", "JPY", "USD"}},
		{"json/alphavantage.json", "AlphaVantage", map[string]interface{}{
			"function":      "CURRENCY_EXCHANGE_RATE",
			"from_currency": "GBP",
			"to_currency":   "EUR",
		}, []string{"Error Message"}},
		{"https://s3.linkpool.io/bridges/cryptocompare.json", "CryptoCompare", nil, []string{"EUR", "JPY", "USD"}},
		{"https://s3.linkpool.io/bridges/alphavantage.json", "AlphaVantage", map[string]interface{}{
			"function":      "CURRENCY_EXCHANGE_RATE",
			"from_currency": "GBP",
			"to_currency":   "EUR",
		}, []string{"Error Message"}},
	}
	for _, c := range cases {
		t.Run(c.path, func(t *testing.T) {
			bs, err := NewJSONBridges(c.path)
			assert.Nil(t, err)
			assert.Len(t, bs, 1)

			for _, b := range bs {
				assert.Equal(t, c.name, b.Opts().Name)
				assert.True(t, b.Opts().Lambda)

				json, err := bridge.ParseInterface(c.req)
				assert.Nil(t, err)
				h := bridge.NewHelper(json)

				obj, err := b.Run(h)
				assert.NotNil(t, obj)
				assert.Nil(t, err)

				resMap, ok := obj.(map[string]interface{})
				assert.True(t, ok)

				for _, k := range c.expectedKeys {
					_, ok := resMap[k]
					assert.True(t, ok)
				}
			}
		})
	}
}
