package response

import (
	"fmt"
)

// Service codes (2 digits: 00-99)
const (
	ServiceCodeCommon   = "00" // Common/General
	ServiceCodeAuth     = "01" // Authentication
	ServiceCodePosts    = "03" // Posts
	ServiceCodeComments = "04" // Comments
)

// Case codes (2 digits: 01-99)
const (
	// Success cases
	CaseCodeSuccess       = "01"
	CaseCodeCreated       = "02"
	CaseCodeUpdated       = "03"
	CaseCodeDeleted       = "04"
	CaseCodeRetrieved     = "05"
	CaseCodeListRetrieved = "06"
	CaseCodeLoginSuccess  = "07"

	// Validation / request errors
	CaseCodeValidationError = "11"
	CaseCodeInvalidFormat   = "13"
	CaseCodeInvalidValue    = "14"
	CaseCodeDuplicateEntry  = "15"

	// Authentication errors
	CaseCodeUnauthorized       = "21"
	CaseCodeInvalidToken       = "22"
	CaseCodeInvalidCredentials = "24"
	CaseCodePermissionDenied   = "27"

	// Not found
	CaseCodeNotFound = "31"

	// Server errors
	CaseCodeInternalError = "55"
)

// BuildResponseCode builds a response code from HTTP status, service code, and case code.
// Format: HTTP_STATUS_CODE (3 digits) + SERVICE_CODE (2 digits) + CASE_CODE (2 digits)
// Example: 4010421 = HTTP 401 + Service 04 + Case 21.
func BuildResponseCode(httpStatus int, serviceCode, caseCode string) int {
	if httpStatus < 100 || httpStatus > 599 {
		panic(fmt.Sprintf("invalid httpStatus: %d", httpStatus))
	}
	if len(serviceCode) != 2 || len(caseCode) != 2 {
		panic("serviceCode and caseCode must be 2 digits")
	}
	for _, r := range serviceCode {
		if r < '0' || r > '9' {
			panic("serviceCode must be numeric")
		}
	}
	for _, r := range caseCode {
		if r < '0' || r > '9' {
			panic("caseCode must be numeric")
		}
	}

	// HTTP (3) + service (2) + case (2)
	codeStr := fmt.Sprintf("%03d%s%s", httpStatus, serviceCode, caseCode)

	code := 0
	for _, char := range codeStr {
		code = code*10 + int(char-'0')
	}
	return code
}

// ParseResponseCode splits a 7-digit code into httpStatus (first 3 digits),
// serviceCode (next 2), caseCode (last 2). Codes not exactly 7 digits return zero values.
func ParseResponseCode(code int) (httpStatus int, serviceCode, caseCode string) {
	if code < 0 || code > 9999999 {
		return 0, "", ""
	}
	codeStr := fmt.Sprintf("%07d", code)
	httpStatus = int(codeStr[0]-'0')*100 + int(codeStr[1]-'0')*10 + int(codeStr[2]-'0')
	serviceCode = codeStr[3:5]
	caseCode = codeStr[5:7]
	return httpStatus, serviceCode, caseCode
}

