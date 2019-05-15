package codec

import "encoding/json"

type jsonCodec struct {}

func (j jsonCodec) Encode(data interface{}) (string, error) {
	s, err := json.Marshal(data)
	return string(s), err
}


func (j jsonCodec) Decode(s string, data interface{}) error {
	return json.Unmarshal([]byte(s), data)
}

func NewJsonCodec() DataCodec {
	return jsonCodec{}
}