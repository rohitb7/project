package protoenumutils

import (
	"fmt"
	protos "www.rvb.com/protos"
)

// String proto Enum helpers for ErrorRetryStatus
func GetErrorRetryStatusEnumValueFromString(stringEnum string) (protos.ErrorRetryStatus, error) {
	if v, ok := protos.ErrorRetryStatus_value[stringEnum]; ok {
		return protos.ErrorRetryStatus(v), nil
	} else {
		return protos.ErrorRetryStatus_NONE, fmt.Errorf("unknown enum value %v for ErrorRetryStatus", stringEnum)
	}
}
