package main

import (
	"github.com/linkpoolio/bridges"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWolframAlpha_Run(t *testing.T) {
	cases := []struct {
		name  string
		data  map[string]interface{}
		error string
	}{
		{"Invalid Index", map[string]interface{}{
			"index": 1,
			"query": "How far is Los Angeles from New York?",
		}, "Invalid index"},
		{"Unauthorised", map[string]interface{}{
			"index": 0,
			"query": "How far is Los Angeles from New York?",
		}, "Unexpected api status code: 403"},
	}
	wa := WolframAlpha{}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			json, err := bridges.ParseInterface(c.data)
			assert.Nil(t, err)

			h := bridges.NewHelper(json)
			_, err = wa.Run(h)

			assert.Equal(t, c.error, err.Error())
		})
	}
}

func TestWolframAlpha_Opts(t *testing.T) {
	cc := WolframAlpha{}
	opts := cc.Opts()
	assert.Equal(t, opts.Name, "WolframAlpha")
	assert.True(t, opts.Lambda)
}
