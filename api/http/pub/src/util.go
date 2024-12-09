package src

import "net/url"

func escapeNonAsciiChars(src string) (dst string) {
	for _, c := range src {
		if c > 127 {
			dst += url.QueryEscape(string(c))
		} else {
			dst += string(c)
		}
	}
	return
}
