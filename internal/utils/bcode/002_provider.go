package bcode

import "net/http"

var (
	ServiceProviderCode = NewBcode(http.StatusOK, 20000, "service provider interface call success")

	ErrServiceProviderBadRequest = NewBcode(http.StatusBadRequest, 20001, "bad request")

	ErrProviderInvalid = NewBcode(http.StatusBadRequest, 20002, "provider invalid")

	ErrProviderIsUnavailable = NewBcode(http.StatusBadRequest, 20003, "service provider is unavailable")

	ErrProviderModelEmpty = NewBcode(http.StatusBadRequest, 20004, "provider model empty")

	ErrProviderUpdateFailed = NewBcode(http.StatusBadRequest, 20005, "provider update failed")

	ErrProviderAuthInfoLost = NewBcode(http.StatusBadRequest, 20006, "provider api auth info lost")

	ErrProviderServiceUrlNotFormat = NewBcode(http.StatusBadRequest, 20007, "provider service url is irregular")
)
