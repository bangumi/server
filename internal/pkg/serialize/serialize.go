package serialize

import (
	"bytes"
	"encoding/json"

	"github.com/trim21/go-phpserialize"
)

func Encode(v any) ([]byte, error) {
	return json.Marshal(v)
}

func Decode(data []byte, v any) error {
	if len(data) == 0 {
		return nil
	}
	if bytes.HasPrefix(data, []byte("a:")) {
		return phpserialize.Unmarshal(data, v)
	}

	return json.Unmarshal(data, v)
}
