package bcode

import "net/http"

var (
	ModelCode = NewBcode(http.StatusOK, 30000, "service interface call success")

	ErrModelBadRequest = NewBcode(http.StatusBadRequest, 30001, " bad request")

	ErrModelIsExist = NewBcode(http.StatusBadRequest, 30002, "provider model already exist")

	ErrModelRecordNotFound = NewBcode(http.StatusBadRequest, 30003, "model not exist")

	ErrAddModel = NewBcode(http.StatusBadRequest, 30004, "model insert db failed")

	ErrDeleteModel = NewBcode(http.StatusBadRequest, 30005, "model delete db failed")

	ErrEngineDeleteModel = NewBcode(http.StatusBadRequest, 30006, "engine delete model failed")

	ErrNoRecommendModel = NewBcode(http.StatusBadRequest, 30007, "No Recommend Model")
)
