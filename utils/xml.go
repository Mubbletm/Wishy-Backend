package utils

import "encoding/xml"

func IsXML(data string) bool {
	return xml.Unmarshal([]byte(data), new(interface{})) == nil
}
