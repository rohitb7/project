package s3_manager

import (
	"context"
	"errors"
	"github.com/aws/smithy-go"
)

var (
	retryErrorCodes = make(map[string]bool)
)

func init() {
	retryErrorCodes["429"] = true //Too Many Requests
	retryErrorCodes["500"] = true //Internal Server Error - "We encountered an internal error. Please try again."
	retryErrorCodes["503"] = true //Service Unavailable - "Reduce your request rate"
	retryErrorCodes["403"] = true //"Service Forbidden - "Access Forbidden"
}

// IsRetryableError - Check if error is retryable
func IsRetryableError(err error) bool {
	if err != nil {
		var ae smithy.APIError
		if errors.As(err, &ae) {
			if _, ok := retryErrorCodes[ae.ErrorCode()]; ok {
				return true
			}
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return true
		}
	}
	return false
}
