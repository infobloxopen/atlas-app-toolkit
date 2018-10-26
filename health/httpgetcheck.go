package health

import (
	"fmt"
	"net/http"
	"time"
)

// HTTPGetCheck returns a Check that performs an HTTP GET request to the
// specified URL. It fails if timeout was exceeded or non-200-OK status code returned.
func HTTPGetCheck(url string, timeout time.Duration) Check {
	client := http.Client{
		Timeout: timeout,
		// never follow redirects
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	return func() error {
		resp, err := client.Get(url)
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
