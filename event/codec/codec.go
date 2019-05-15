package codec

type DataEncoder interface {
	Encode(data interface{}) (string, error)
}

type DataDecoder interface {
	Decode(str string, data interface{}) error
}


type DataCodec interface {
	Encode(data interface{}) (string, error)
	Decode(str string, data interface{}) error
}