package http

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"fmt"
	"io"
)

func readBody(o *requestOption, encoding string, body io.Reader) ([]byte, error) {
	o.debugf("use encoding '%s'", encoding)

	switch encoding {
	case "":
		// do nothing
	case "gzip":
		reader, err := gzip.NewReader(body)
		if err != nil {
			return nil, fmt.Errorf("gzip.NewReader error (%w)", err)
		}
		body = reader
		defer reader.Close()

	case "deflate":
		reader := flate.NewReader(body)
		body = reader
		defer reader.Close()

	default:
		return nil, fmt.Errorf("unsupported encoding: %s", encoding)
	}

	if o.progressCB == nil {
		b, err := io.ReadAll(body)
		if err != nil {
			return nil, fmt.Errorf("io.ReadAll error (%w)", err)
		}
		return b, nil
	}

	buff := &bytes.Buffer{}
	w := io.MultiWriter(buff, o.progress)
	if _, err := io.Copy(w, body); err != nil {
		return nil, fmt.Errorf("io.ReadAll error (%w)", err)
	}

	return buff.Bytes(), nil
}
