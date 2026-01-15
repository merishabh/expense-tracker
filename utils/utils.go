package utils

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
)

func DecodeBase64URL(data string) string {
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

func StripHTMLTags(html string) string {
	re := regexp.MustCompile("<[^>]*>")
	text := re.ReplaceAllString(html, "")
	// Normalize whitespace - replace multiple spaces/newlines with single space
	re = regexp.MustCompile(`\s+`)
	text = re.ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}
