package main

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
)

func decodeBase64URL(data string) string {
	data = strings.ReplaceAll(data, "-", "+")
	data = strings.ReplaceAll(data, "_", "/")
	switch len(data) % 4 {
	case 2:
		data += "=="
	case 3:
		data += "="
	}

	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		fmt.Println("Error decoding body:", err)
		return ""
	}
	return string(decoded)
}

func stripHTMLTags(html string) string {
	re := regexp.MustCompile("<[^>]*>")
	return re.ReplaceAllString(html, "")
}
