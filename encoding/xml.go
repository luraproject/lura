package encoding

import (
	"io"

	"github.com/clbanning/mxj"
)

// XML is the key for the xml encoding
const XML = "xml"

// NewXMLDecoder return the right XML decoder
func NewXMLDecoder(isCollection bool) Decoder {
	if isCollection {
		return XMLCollectionDecoder
	}
	return XMLDecoder
}

// XMLDecoder implements the Decoder interface
func XMLDecoder(r io.Reader, v *map[string]interface{}) error {
	mv, err := mxj.NewMapXmlReader(r)
	if err != nil {
		return err
	}
	*v = mv
	return nil

}

// XMLCollectionDecoder implements the Decoder interface over a collection
func XMLCollectionDecoder(r io.Reader, v *map[string]interface{}) error {
	mv, err := mxj.NewMapXmlReader(r)
	if err != nil {
		return err
	}
	*(v) = map[string]interface{}{"collection": mv}
	return nil
}
