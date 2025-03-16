package bcode

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"intel.com/aog/internal/datastore"
)

// Error Code of AOG contains 5 digits, the first 3 digits should be reversed and indicates the category of concept
// the last two digits indicates the error number
// For example, business code 11001 should split to 110 and 01, it means the code belongs to the 011 category env, and it's the 01 number error.

// SuccessCode a success code
var SuccessCode = NewBcode(200, 200, "success")

// ErrServer an unexpected mistake.
var ErrServer = NewBcode(500, 500, "The service has lapsed.")

// ErrForbidden check user perms failure
var ErrForbidden = NewBcode(403, 403, "403 Forbidden")

// ErrUnauthorized check user auth failure
var ErrUnauthorized = NewBcode(401, 401, "401 Unauthorized")

// ErrNotFound the request resource is not found
var ErrNotFound = NewBcode(404, 404, "404 Not Found")

// ErrUpstreamNotFound the proxy upstream is not found
var ErrUpstreamNotFound = NewBcode(502, 502, "Upstream not found")

// Bcode business error code
type Bcode struct {
	HTTPCode     int32  `json:"-"`
	BusinessCode int32  `json:"business_code"`
	Message      string `json:"message"`
}

func (b *Bcode) Error() string {
	switch {
	case b.Message != "":
		return b.Message
	default:
		return "something went wrong, please see the aog server logs for details"
	}
}

// SetMessage set new message and return a new error instance
func (b *Bcode) SetMessage(message string) *Bcode {
	return &Bcode{
		HTTPCode:     b.HTTPCode,
		BusinessCode: b.BusinessCode,
		Message:      message,
	}
}

var bcodeMap map[int32]*Bcode

// NewBcode new error code
func NewBcode(httpCode, businessCode int32, message string) *Bcode {
	if bcodeMap == nil {
		bcodeMap = make(map[int32]*Bcode)
	}
	if _, exit := bcodeMap[businessCode]; exit {
		panic("error business code is exist")
	}
	bcode := &Bcode{HTTPCode: httpCode, BusinessCode: businessCode, Message: message}
	bcodeMap[businessCode] = bcode
	return bcode
}

// ReturnHTTPError Unified handling of all types of errors, generating a standard return structure.
func ReturnHTTPError(c *gin.Context, err error) {
	c.SetAccepted(gin.MIMEJSON)
	ReturnError(c, err)
}

// ReturnError Unified handling of all types of errors, generating a standard return structure.
func ReturnError(c *gin.Context, err error) {
	var bcode *Bcode
	if errors.As(err, &bcode) {
		c.JSON(int(bcode.HTTPCode), err)
		return
	}

	if errors.Is(err, datastore.ErrRecordNotExist) {
		c.JSON(http.StatusNotFound, err)
		return
	}

	var validErr validator.ValidationErrors
	if errors.As(err, &validErr) {
		c.JSON(http.StatusBadRequest, Bcode{
			HTTPCode:     http.StatusBadRequest,
			BusinessCode: 400,
			Message:      err.Error(),
		})
		return
	}

	c.JSON(http.StatusInternalServerError, Bcode{
		HTTPCode:     http.StatusInternalServerError,
		BusinessCode: 500,
		Message:      err.Error(),
	})
}
