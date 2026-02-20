package api

import (
	"context"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
)

// customCheckRetry mirrors the Python custom_retry_predicate:
//   - 400 (Bad Request) and 404 (Not Found) → do not retry
//   - 500, 502, 503, 504 and network errors  → retry (via DefaultRetryPolicy)
func customCheckRetry(ctx context.Context, resp *http.Response, err error) (bool, error) {
	if resp != nil {
		switch resp.StatusCode {
		case http.StatusBadRequest, http.StatusNotFound:
			return false, nil
		}
	}
	return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
}
