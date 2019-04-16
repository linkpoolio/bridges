package main

import (
	"bytes"
	"encoding/json"
	"github.com/linkpoolio/bridges/bridge"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
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
		{"json/placeholder.json", "Placeholder", map[string]interface{}{
			"page": 1,
		}, []string{"page", "per_page", "total_pages"}},
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

func TestNewJSONBridges_Errors(t *testing.T) {
	_, err := NewJSONBridges("")
	assert.Equal(t, "Empty bridge URI given", err.Error())

	_, err = NewJSONBridges("json/invalid.json")
	assert.Equal(t, "unexpected end of JSON input", err.Error())

	_, err = NewJSONBridges("http://invalidqwerty.com")
	assert.Contains(t, err.Error(), "no such host")
}

func TestHandler(t *testing.T) {
	p := map[string]interface{}{
		"jobRunId": "1234",
	}
	pb, err := json.Marshal(p)
	assert.Nil(t, err)

	req, err := http.NewRequest(http.MethodPost, "/", bytes.NewReader(pb))
	assert.Nil(t, err)
	rr := httptest.NewRecorder()

	err = os.Setenv("BRIDGE", "json/cryptocompare.json")
	assert.Nil(t, err)
	Handler(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	err = os.Unsetenv("BRIDGE")
	assert.Nil(t, err)

	body, err := ioutil.ReadAll(rr.Body)
	assert.Nil(t, err)
	json, err := bridge.Parse(body)
	assert.Nil(t, err)

	assert.Equal(t, "1234", json.Get("jobRunId").String())

	data := json.Get("data").Map()
	_, ok := data["USD"]
	assert.True(t, ok)
	_, ok = data["JPY"]
	assert.True(t, ok)
	_, ok = data["EUR"]
	assert.True(t, ok)
}

func TestHandler_NilBridge(t *testing.T) {
	p := map[string]interface{}{
		"jobRunId": "1234",
	}
	pb, err := json.Marshal(p)
	assert.Nil(t, err)

	req, err := http.NewRequest(http.MethodPost, "/", bytes.NewReader(pb))
	assert.Nil(t, err)
	rr := httptest.NewRecorder()

	err = os.Setenv("BRIDGE", "")
	assert.Nil(t, err)
	Handler(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	err = os.Unsetenv("BRIDGE")
	assert.Nil(t, err)

	body, err := ioutil.ReadAll(rr.Body)
	assert.Nil(t, err)
	assert.Equal(t, "No bridge set", string(body))
}

func TestHandler_InvalidBridge(t *testing.T) {
	p := map[string]interface{}{
		"jobRunId": "1234",
	}
	pb, err := json.Marshal(p)
	assert.Nil(t, err)

	req, err := http.NewRequest(http.MethodPost, "/", bytes.NewReader(pb))
	assert.Nil(t, err)
	rr := httptest.NewRecorder()

	err = os.Setenv("BRIDGE", "json/invalid.json")
	assert.Nil(t, err)
	Handler(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	err = os.Unsetenv("BRIDGE")
	assert.Nil(t, err)

	body, err := ioutil.ReadAll(rr.Body)
	assert.Nil(t, err)
	assert.Equal(t, "unexpected end of JSON input", string(body))
}