package cache

import (
	"crypto/md5"
	"crypto/sha1"
	"fmt"

	"github.com/kataras/iris"
)

// RequestPathToMD5 converts the request's path to a md5 hash
func RequestPathToMD5(ctx iris.Context) string {
	u := ctx.Request().RequestURI
	return fmt.Sprintf("%x", md5.Sum([]byte(u))) // or ctx.Path if no subdomains involved.
}

// RequestPathToSha1 converts the request's path to a sha1 hash
func RequestPathToSha1(ctx iris.Context) string {
	u := ctx.Request().RequestURI
	return fmt.Sprintf("%x", sha1.Sum([]byte(u)))
}
