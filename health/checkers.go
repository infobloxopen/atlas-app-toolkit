package health

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
)

func HttpGetCheck(url *url.URL, timeout time.Duration) Check {
	client := http.Client{
		Timeout: timeout,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	return func() error {
		resp, err := client.Get(url.String())
		if err != nil {
			return err
		}
		resp.Body.Close()
		if resp.StatusCode != 200 {
			return fmt.Errorf("%d: %s", resp.StatusCode, resp.Status)
		}
		return nil
	}
}
