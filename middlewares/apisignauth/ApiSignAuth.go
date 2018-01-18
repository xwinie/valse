package apisignauth

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/valyala/fasthttp"
	"github.com/xwinie/valse"
)

//Config 需要的结构配置
type Config struct {
	f       func(string) string
	timeout int
}

//APISignAuth api sign auth
func APISignAuth(c Config, next valse.RequestHandler) valse.RequestHandler {
	return func(ctx *valse.Context) error {
		if string(ctx.HeaderParameter("appid")) == "" {
			return valse.NewHTTPMessage(http.StatusForbidden, "miss appid header")
		}
		appsecret := c.f(string(ctx.HeaderParameter("appid")))
		if appsecret == "" {
			return valse.NewHTTPMessage(fasthttp.StatusForbidden, "not exist this appid")
		}
		clientSignature := string(ctx.HeaderParameter("signature"))
		if clientSignature == "" {
			return valse.NewHTTPMessage(fasthttp.StatusForbidden, "miss signature header")
		}
		if string(ctx.HeaderParameter("timestamp")) == "" {
			return valse.NewHTTPMessage(fasthttp.StatusForbidden, "miss timestamp header")
		}
		u, err := time.Parse("2006-01-02 15:04:05", string(ctx.HeaderParameter("timestamp")))
		if err != nil {

			return valse.NewHTTPMessage(fasthttp.StatusForbidden, "timestamp format is error, should 2006-01-02 15:04:05")
		}
		t := time.Now()
		if t.Sub(u).Seconds() > float64(c.timeout) {
			return valse.NewHTTPMessage(fasthttp.StatusForbidden, "timeout! the request time is long ago, please try again")
		}
		var requestURL string
		var boy []byte
		if ctx.IsGet() {
			requestURL = string(bytes.TrimLeft(ctx.RequestURI(), "/"))
		} else {
			boy = ctx.GetBody()
		}
		serviceSignature := Signature(appsecret, string(ctx.Method()), boy, requestURL, string(ctx.HeaderParameter("timestamp")))
		if clientSignature != serviceSignature {
			return valse.NewHTTPMessage(fasthttp.StatusForbidden, "Signature Failed")
		}
		return next(ctx)
	}

}

// Signature used to generate signature with the appsecret/method/params/RequestURI
func Signature(appSecret, method string, body []byte, RequestURL string, timestamp string) (result string) {
	stringToSign := fmt.Sprintf("%v\n%v\n%v\n%v\n", method, string(body), RequestURL, timestamp)
	fmt.Println(stringToSign)
	sha256 := sha256.New
	hash := hmac.New(sha256, []byte(appSecret))
	hash.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}
