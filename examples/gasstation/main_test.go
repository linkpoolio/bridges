package main

import (
	"github.com/linkpoolio/bridges"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGasStation_Run(t *testing.T) {
	wa := GasStation{}

	h := bridges.NewHelper(&bridges.JSON{})
	val, err := wa.Run(h)

	avg, ok := val.(float64)
	assert.True(t, ok)
	assert.True(t, avg > 0)
	assert.Nil(t, err)
}

func TestGasStation_Opts(t *testing.T) {
	cc := GasStation{}
	opts := cc.Opts()
	assert.Equal(t, opts.Name, "GasStation")
	assert.True(t, opts.Lambda)
}

