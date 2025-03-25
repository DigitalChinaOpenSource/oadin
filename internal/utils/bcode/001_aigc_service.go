package bcode

import "net/http"

var (
	AIGCServiceCode = NewBcode(http.StatusOK, 10000, "service interface call success")

	ErrAIGCServiceBadRequest = NewBcode(http.StatusBadRequest, 10001, "bad request")

	ErrAIGCServiceInitEnv = NewBcode(http.StatusBadRequest, 10002, "set env failed")

	ErrAIGCServiceInstallEngine = NewBcode(http.StatusBadRequest, 10003, "install model engine failed")

	ErrAIGCServiceStartEngine = NewBcode(http.StatusBadRequest, 10004, "start model engine failed")

	ErrGetEngineModelList = NewBcode(http.StatusBadRequest, 10005, "get model list failed")

	ErrEnginePullModel = NewBcode(http.StatusBadRequest, 10006, "pull model failed")

	ErrAIGCServiceAddProvider = NewBcode(http.StatusBadRequest, 10007, "provider insert db failed")

	ErrAIGCServiceProviderIsExist = NewBcode(http.StatusBadRequest, 10009, "provider already exist")

	ErrServiceRecordNotFound = NewBcode(http.StatusBadRequest, 10011, "service not found")

	ErrServiceUpdateFailed = NewBcode(http.StatusBadRequest, 10012, "service edit failed")

	ErrAddModelService = NewBcode(http.StatusBadRequest, 10013, "add model service failed")

	ErrAIGCServiceVersionNotMatch = NewBcode(http.StatusBadRequest, 10014, "byze version not match")

	ErrUnSupportAIGCService = NewBcode(http.StatusBadRequest, 10015, "unsupport byze service")

	ErrUnSupportHybridPolicy = NewBcode(http.StatusBadRequest, 10016, "unsupport hybrid policy")

	ErrUnSupportFlavor = NewBcode(http.StatusBadRequest, 10017, "unsupport api flavor")

	ErrUnSupportAuthType = NewBcode(http.StatusBadRequest, 10018, "unsupport auth type")
)
