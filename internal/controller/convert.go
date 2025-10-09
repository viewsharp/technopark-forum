package controller

import (
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

// Helper functions to create pointers
func ptrString(s string) *string {
	return &s
}

func ptrBool(b bool) *bool {
	return &b
}

// Convert int32 pointer to int64 pointer
func int32ToInt64Ptr(i *int32) *int64 {
	if i == nil {
		return nil
	}
	val := int64(*i)
	return &val
}

// Convert openapi_types.Email to string
func emailToString(email openapi_types.Email) string {
	return string(email)
}

// Convert string to openapi_types.Email
func stringToEmail(s string) openapi_types.Email {
	return openapi_types.Email(s)
}

// Convert time.Time pointer to string
func timePtrToString(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339Nano)
}

// Convert int64 to int
func int64ToInt(i int64) int {
	return int(i)
}

// Convert int32 to int
func int32ToInt(i int32) int {
	return int(i)
}

// Convert int32 pointer to int64 pointer for Post IDs
func int32PtrToInt64Ptr(i *int32) *int64 {
	if i == nil {
		return nil
	}
	val := int64(*i)
	return &val
}

// Convert int64 pointer to int32 pointer for Post IDs
func int64PtrToInt32Ptr(i *int64) *int32 {
	if i == nil {
		return nil
	}
	val := int32(*i)
	return &val
}
