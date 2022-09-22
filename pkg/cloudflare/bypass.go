// from https://github.com/trewisscotch/CloudFlare-ByPass-Go./blob/669ea5cec99e3030d7984873a431d9f6335a3f48/round_tripper.go
package cloudflare

import (
	"crypto/tls"
	"net/http"
	"net/http/cookiejar"

	"github.com/sirupsen/logrus"
)

const userAgent = `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/101.0.4951.64 Safari/537.36`

type Transport struct {
	Logger   *logrus.Logger
	upstream http.RoundTripper
	Cookies  http.CookieJar
}

func NewClient(logger *logrus.Logger) (c *http.Client, err error) {

	scraper_transport, err := NewTransport(logger, &http.Transport{
		TLSClientConfig: &tls.Config{
			PreferServerCipherSuites: false,
			CurvePreferences:         []tls.CurveID{tls.CurveP256, tls.CurveP384, tls.CurveP521, tls.X25519},
		},
	})
	if err != nil {
		return
	}

	c = &http.Client{
		Transport: scraper_transport,
		Jar:       scraper_transport.Cookies,
	}

	return
}

func NewTransport(logger *logrus.Logger, upstream http.RoundTripper) (*Transport, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	return &Transport{logger, upstream, jar}, nil
}

func (t Transport) RoundTrip(r *http.Request) (*http.Response, error) {
	defaultHeader := map[string]string{
		"Connection":                "keep-alive",
		"Upgrade-Insecure-Requests": "1",
		"User-Agent":                userAgent,
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
		"Accept-Language":           "en-US,en;q=0.9",
		"Accept-Encoding":           "gzip, deflate, br",
	}
	for key, value := range defaultHeader {
		if r.Header.Get(key) == "" {
			r.Header.Set(key, value)

		}
	}
	resp, err := t.upstream.RoundTrip(r)
	if err != nil {
		return nil, err
	}

	// Check if Cloudflare anti-bot is on
	// t.Logger.Debugln("cloudflare", resp.StatusCode)
	// if (resp.StatusCode == 503 || resp.StatusCode == 403) && strings.HasPrefix(resp.Header.Get("Server"), "cloudflare") {
	// 	t.Logger.Printf("Solving challenge for %s", resp.Request.URL.Hostname())
	// 	resp, err := t.solveChallenge(resp)

	// 	return resp, err
	// }

	// https://zc310.tech/blog/post/my/2020/golang-http-brotli/
	// https://medium.com/sellerapp/decoding-brotli-in-golang-e024dcf34be1
	// switch resp.Header.Get(headers.ContentEncoding) {
	// case "br":
	// 	return io.ReadAll(cbrotli.NewReader(resp.Body))
	// case "gzip":
	// 	gr, err := gzip.NewReader(resp.Body)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	return io.ReadAll(gr)
	// case "deflate":
	// 	zr := flate.NewReader(resp.Body)
	// 	defer zr.Close()
	// 	return io.ReadAll(zr)

	// default:
	// 	return io.ReadAll(resp.Body)
	// }
	return resp, err
}
