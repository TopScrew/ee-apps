// Code generated by goa v3.14.1, DO NOT EDIT.
//
// ks3 HTTP server encoders and decoders
//
// Command:
// $ goa gen github.com/PingCAP-QE/ee-apps/dl/design

package server

import (
	"context"
	"net/http"
	"strconv"

	ks3 "github.com/PingCAP-QE/ee-apps/dl/gen/ks3"
	goahttp "goa.design/goa/v3/http"
)

// EncodeDownloadObjectResponse returns an encoder for responses returned by
// the ks3 download-object endpoint.
func EncodeDownloadObjectResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, any) error {
	return func(ctx context.Context, w http.ResponseWriter, v any) error {
		res, _ := v.(*ks3.DownloadObjectResult)
		ctx = context.WithValue(ctx, goahttp.ContentTypeKey, "application/octet-stream")
		{
			val := res.Length
			lengths := strconv.FormatInt(val, 10)
			w.Header().Set("Content-Length", lengths)
		}
		w.Header().Set("Content-Disposition", res.ContentDisposition)
		w.WriteHeader(http.StatusOK)
		return nil
	}
}

// DecodeDownloadObjectRequest returns a decoder for requests sent to the ks3
// download-object endpoint.
func DecodeDownloadObjectRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (any, error) {
	return func(r *http.Request) (any, error) {
		var (
			bucket string
			key    string

			params = mux.Vars(r)
		)
		bucket = params["bucket"]
		key = params["key"]
		payload := NewDownloadObjectPayload(bucket, key)

		return payload, nil
	}
}
