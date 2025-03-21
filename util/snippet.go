package util

import (
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"github.com/microcosm-cc/bluemonday"
	"regexp"
	"unicode/utf8"
)

var attrNamesInclude = map[string]bool{
	"articlebody":       true,
	"articlesection":    true,
	"caption":           true,
	"categories":        true,
	"citation":          true,
	"description":       true,
	"headline":          true,
	"jobtitle":          true,
	"keywords":          true,
	"name":              true,
	"object":            true,
	"sourcecategories":  true,
	"sourcedescription": true,
	"sourceimagetitle":  true,
	"sourcetitle":       true,
	"subject":           true,
	"summary":           true,
	"title":             true,
}

var htmlPolicy = bluemonday.
						StrictPolicy().
						AddSpaceWhenStrippingTag(true)
var reMultiSpace = regexp.MustCompile(`\s+`) // Match one or more whitespace characters

func ExtractTextSnippet(evt *pb.CloudEvent, lenMax int) (snippet string) {
	snippet = htmlPolicy.Sanitize(evt.GetTextData())
	snippet = reMultiSpace.ReplaceAllString(snippet, " ")
	if len(snippet) < lenMax {
		for k, v := range evt.Attributes {
			if !attrNamesInclude[k] {
				continue
			}
			vs := htmlPolicy.Sanitize(v.GetCeString())
			vs = reMultiSpace.ReplaceAllString(vs, " ")
			if vs == "" {
				continue
			}
			snippet += "\n" + vs
			if len(snippet) > lenMax {
				break
			}
		}
	}
	if len(snippet) > lenMax {
		snippet = TruncateUTF8(snippet, lenMax)
	}
	return
}

func TruncateUTF8(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}

	// Trim to maxBytes, but ensure we don't cut a multi-byte character
	truncated := s[:maxBytes]

	// Check if we accidentally cut a UTF-8 character
	for !utf8.ValidString(truncated) {
		_, size := utf8.DecodeLastRuneInString(truncated)
		truncated = truncated[:len(truncated)-size] // Remove last incomplete rune
	}

	return truncated
}
