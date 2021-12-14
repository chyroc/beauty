package helper

import (
	"regexp"

	"github.com/chyroc/gorequests"
)

func GetOneMatchString(s string, regexp *regexp.Regexp) string {
	m := regexp.FindStringSubmatch(s)
	if len(m) == 2 {
		return m[1]
	}
	return ""
}

var Request *gorequests.Factory

func init() {
	Request = gorequests.NewFactory(
		gorequests.WithLogger(gorequests.NewDiscardLogger()),
		gorequests.WithHeader("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/93.0.4577.82 Safari/537.36"),
	)
}
