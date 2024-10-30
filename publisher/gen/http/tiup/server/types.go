// Code generated by goa v3.19.1, DO NOT EDIT.
//
// tiup HTTP server types
//
// Command:
// $ goa gen github.com/PingCAP-QE/ee-apps/publisher/design

package server

import (
	tiup "github.com/PingCAP-QE/ee-apps/publisher/gen/tiup"
	goa "goa.design/goa/v3/pkg"
)

// RequestToPublishRequestBody is the type of the "tiup" service
// "request-to-publish" endpoint HTTP request body.
type RequestToPublishRequestBody struct {
	// The full url of the pushed OCI artifact, contain the tag part. It will parse
	// the repo from it.
	ArtifactURL *string `form:"artifact_url,omitempty" json:"artifact_url,omitempty" xml:"artifact_url,omitempty"`
	// Force set the version. Default is the artifact version read from
	// `org.opencontainers.image.version` of the manifest config.
	Version *string `form:"version,omitempty" json:"version,omitempty" xml:"version,omitempty"`
	// Staging is http://tiup.pingcap.net:8988, product is
	// http://tiup.pingcap.net:8987.
	TiupMirror *string `form:"tiup-mirror,omitempty" json:"tiup-mirror,omitempty" xml:"tiup-mirror,omitempty"`
	// The request id
	RequestID *string `form:"request_id,omitempty" json:"request_id,omitempty" xml:"request_id,omitempty"`
}

// NewRequestToPublishPayload builds a tiup service request-to-publish endpoint
// payload.
func NewRequestToPublishPayload(body *RequestToPublishRequestBody) *tiup.RequestToPublishPayload {
	v := &tiup.RequestToPublishPayload{
		ArtifactURL: *body.ArtifactURL,
		Version:     body.Version,
		TiupMirror:  *body.TiupMirror,
		RequestID:   body.RequestID,
	}

	return v
}

// NewQueryPublishingStatusPayload builds a tiup service
// query-publishing-status endpoint payload.
func NewQueryPublishingStatusPayload(requestID string) *tiup.QueryPublishingStatusPayload {
	v := &tiup.QueryPublishingStatusPayload{}
	v.RequestID = requestID

	return v
}

// ValidateRequestToPublishRequestBody runs the validations defined on
// Request-To-PublishRequestBody
func ValidateRequestToPublishRequestBody(body *RequestToPublishRequestBody) (err error) {
	if body.ArtifactURL == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("artifact_url", "body"))
	}
	if body.TiupMirror == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("tiup-mirror", "body"))
	}
	return
}
