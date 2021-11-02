package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseArgs(t *testing.T) {

	c1 := map[byte]byte{
		0x01: 0x05,
		0x02: 0x06,
		0x06: 0x00,
	}
	e1 := []byte{0x3, 0x2, 0x6, 0x1, 0x5, 0x6, 0x0}
	require.Equal(t, e1, classificationToByteArray(c1))
}
