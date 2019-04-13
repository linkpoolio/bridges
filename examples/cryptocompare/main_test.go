package main

import (
	"github.com/jleeh/bridges/bridge"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCryptoCompare_Run(t *testing.T) {
	cc := CryptoCompare{}
	obj, err := cc.Run(bridge.NewHelper(nil))
	assert.Nil(t, err)

	resp, ok := obj.(map[string]interface{})
	assert.True(t, ok)

	_, ok = resp["USD"]
	assert.True(t, ok)
	_, ok = resp["JPY"]
	assert.True(t, ok)
	_, ok = resp["EUR"]
	assert.True(t, ok)
}

func TestCryptoCompare_Opts(t *testing.T) {
	cc := CryptoCompare{}
	opts := cc.Opts()
	assert.Equal(t, opts.Name, "CryptoCompare")
	assert.True(t, opts.Lambda)
}
