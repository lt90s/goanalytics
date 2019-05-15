package codec

import (
	"github.com/stretchr/testify/require"
	"testing"
)

type fooBar struct {
	Foo int `json:"foo"`
	Bar string `json:"bar"`
}

var fb = fooBar{
	Foo: 1,
	Bar: "bar",
}

var fbs = `{"foo":1,"bar":"bar"}`

func TestJsonCodec_Encode(t *testing.T) {
	codec := NewJsonCodec()
	s, err := codec.Encode(fb)
	require.NoError(t, err)
	require.Equal(t, fbs, s)
}

func TestJsonCodec_Decode(t *testing.T) {
	codec := NewJsonCodec()
	var data fooBar
	err := codec.Decode(fbs, &data)
	require.NoError(t, err)
	require.Equal(t, fb, data)
}