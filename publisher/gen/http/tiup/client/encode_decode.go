// Code generated by goa v3.19.1, DO NOT EDIT.
//
// tiup HTTP client encoders and decoders
//
// Command:
// $ goa gen github.com/PingCAP-QE/ee-apps/publisher/design

package client

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"

	tiup "github.com/PingCAP-QE/ee-apps/publisher/gen/tiup"
	goahttp "goa.design/goa/v3/http"
	goa "goa.design/goa/v3/pkg"
)

// BuildRequestToPublishRequest instantiates a HTTP request object with method
// and path set to call the "tiup" service "request-to-publish" endpoint
func (c *Client) BuildRequestToPublishRequest(ctx context.Context, v any) (*http.Request, error) {
	u := &url.URL{Scheme: c.scheme, Host: c.host, Path: RequestToPublishTiupPath()}
	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		return nil, goahttp.ErrInvalidURL("tiup", "request-to-publish", u.String(), err)
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}

	return req, nil
}

// EncodeRequestToPublishRequest returns an encoder for requests sent to the
// tiup request-to-publish server.
func EncodeRequestToPublishRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, any) error {
	return func(req *http.Request, v any) error {
		p, ok := v.(*tiup.RequestToPublishPayload)
		if !ok {
			return goahttp.ErrInvalidType("tiup", "request-to-publish", "*tiup.RequestToPublishPayload", v)
		}
		body := NewRequestToPublishRequestBody(p)
		if err := encoder(req).Encode(&body); err != nil {
			return goahttp.ErrEncodingError("tiup", "request-to-publish", err)
		}
		return nil
	}
}

// DecodeRequestToPublishResponse returns a decoder for responses returned by
// the tiup request-to-publish endpoint. restoreBody controls whether the
// response body should be restored after having been read.
func DecodeRequestToPublishResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (any, error) {
	return func(resp *http.Response) (any, error) {
		if restoreBody {
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			resp.Body = io.NopCloser(bytes.NewBuffer(b))
			defer func() {
				resp.Body = io.NopCloser(bytes.NewBuffer(b))
			}()
		} else {
			defer resp.Body.Close()
		}
		switch resp.StatusCode {
		case http.StatusOK:
			var (
				body []string
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("tiup", "request-to-publish", err)
			}
			return body, nil
		default:
			body, _ := io.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("tiup", "request-to-publish", resp.StatusCode, string(body))
		}
	}
}

// BuildQueryPublishingStatusRequest instantiates a HTTP request object with
// method and path set to call the "tiup" service "query-publishing-status"
// endpoint
func (c *Client) BuildQueryPublishingStatusRequest(ctx context.Context, v any) (*http.Request, error) {
	var (
		requestID string
	)
	{
		p, ok := v.(*tiup.QueryPublishingStatusPayload)
		if !ok {
			return nil, goahttp.ErrInvalidType("tiup", "query-publishing-status", "*tiup.QueryPublishingStatusPayload", v)
		}
		requestID = p.RequestID
	}
	u := &url.URL{Scheme: c.scheme, Host: c.host, Path: QueryPublishingStatusTiupPath(requestID)}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, goahttp.ErrInvalidURL("tiup", "query-publishing-status", u.String(), err)
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}

	return req, nil
}

// DecodeQueryPublishingStatusResponse returns a decoder for responses returned
// by the tiup query-publishing-status endpoint. restoreBody controls whether
// the response body should be restored after having been read.
func DecodeQueryPublishingStatusResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (any, error) {
	return func(resp *http.Response) (any, error) {
		if restoreBody {
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			resp.Body = io.NopCloser(bytes.NewBuffer(b))
			defer func() {
				resp.Body = io.NopCloser(bytes.NewBuffer(b))
			}()
		} else {
			defer resp.Body.Close()
		}
		switch resp.StatusCode {
		case http.StatusOK:
			var (
				body string
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("tiup", "query-publishing-status", err)
			}
			if !(body == "queued" || body == "processing" || body == "success" || body == "failed" || body == "canceled") {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError("body", body, []any{"queued", "processing", "success", "failed", "canceled"}))
			}
			if err != nil {
				return nil, goahttp.ErrValidationError("tiup", "query-publishing-status", err)
			}
			return body, nil
		default:
			body, _ := io.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("tiup", "query-publishing-status", resp.StatusCode, string(body))
		}
	}
}