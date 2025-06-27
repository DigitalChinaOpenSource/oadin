package bcode

import "net/http"

var (
	AIGCServiceCode = NewBcode(http.StatusOK, 10000, "service interface call success")

	ErrAIGCServiceBadRequest = NewBcode(http.StatusBadRequest, 10001, "bad request")

	ErrAIGCServiceInitEnv = NewBcode(http.StatusInternalServerError, 10002, "set env failed")

	ErrAIGCServiceInstallEngine = NewBcode(http.StatusInternalServerError, 10003, "install model engine failed")

	ErrAIGCServiceStartEngine = NewBcode(http.StatusInternalServerError, 10004, "start model engine failed")

	ErrGetEngineModelList = NewBcode(http.StatusInternalServerError, 10005, "get model list failed")

	ErrEnginePullModel = NewBcode(http.StatusInternalServerError, 10006, "pull model failed")

	ErrAIGCServiceAddProvider = NewBcode(http.StatusInternalServerError, 10007, "provider insert db failed")

	ErrAIGCServiceProviderIsExist = NewBcode(http.StatusConflict, 10009, "provider already exist")

	ErrServiceRecordNotFound = NewBcode(http.StatusNotFound, 10011, "service not found")

	ErrServiceUpdateFailed = NewBcode(http.StatusInternalServerError, 10012, "service edit failed")

	ErrAddModelService = NewBcode(http.StatusInternalServerError, 10013, "add model service failed")

	ErrAIGCServiceVersionNotMatch = NewBcode(http.StatusUnprocessableEntity, 10014, "oadin version not match")

	ErrUnSupportAIGCService = NewBcode(http.StatusBadRequest, 10015, "unsupport oadin service")

	ErrUnSupportHybridPolicy = NewBcode(http.StatusBadRequest, 10016, "unsupport hybrid policy")

	ErrUnSupportFlavor = NewBcode(http.StatusBadRequest, 10017, "unsupport api flavor")

	ErrUnSupportAuthType = NewBcode(http.StatusBadRequest, 10018, "unsupport auth type")
)
