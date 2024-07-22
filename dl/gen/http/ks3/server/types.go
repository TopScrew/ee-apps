// Code generated by goa v3.16.1, DO NOT EDIT.
//
// ks3 HTTP server types
//
// Command:
// $ goa gen github.com/PingCAP-QE/ee-apps/dl/design

package server

import (
	ks3 "github.com/PingCAP-QE/ee-apps/dl/gen/ks3"
)

// NewDownloadObjectPayload builds a ks3 service download-object endpoint
// payload.
func NewDownloadObjectPayload(bucket string, key string) *ks3.DownloadObjectPayload {
	v := &ks3.DownloadObjectPayload{}
	v.Bucket = bucket
	v.Key = key

	return v
}
