// +build !go1.8

package fasthttptreemux

import "net/url"

func unescape(path string) (string, error) {
	return url.QueryUnescape(path)
}
