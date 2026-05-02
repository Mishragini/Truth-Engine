package utils

import (
	"io"
	"strings"

	"golang.org/x/net/html"
)

func Sanitize(str string) (string, error) {
	r := strings.NewReader(str)
	z := html.NewTokenizer(r)
	var htmlText string

loop:
	for {
		tt := z.Next()

		switch tt {
		case html.ErrorToken:
			if z.Err() == io.EOF {
				break loop
			}
			return "", z.Err()
		case html.TextToken:
			htmlText += z.Token().Data
		}
	}

	return html.UnescapeString(htmlText), nil
}
