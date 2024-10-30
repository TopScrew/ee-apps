// Code generated by goa v3.19.1, DO NOT EDIT.
//
// fileserver HTTP client types
//
// Command:
// $ goa gen github.com/PingCAP-QE/ee-apps/publisher/design

package client

import (
	fileserver "github.com/PingCAP-QE/ee-apps/publisher/gen/fileserver"
)

// RequestToPublishRequestBody is the type of the "fileserver" service
// "request-to-publish" endpoint HTTP request body.
type RequestToPublishRequestBody struct {
	// The full url of the pushed OCI artifact, contain the tag part. It will parse
	// the repo from it.
	ArtifactURL string `form:"artifact_url" json:"artifact_url" xml:"artifact_url"`
}

// NewRequestToPublishRequestBody builds the HTTP request body from the payload
// of the "request-to-publish" endpoint of the "fileserver" service.
func NewRequestToPublishRequestBody(p *fileserver.RequestToPublishPayload) *RequestToPublishRequestBody {
	body := &RequestToPublishRequestBody{
		ArtifactURL: p.ArtifactURL,
	}
	return body
}
