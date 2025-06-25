package bcode

import "net/http"

var (
	ModelCode = NewBcode(http.StatusOK, 30000, "service interface call success")

	ErrModelBadRequest = NewBcode(http.StatusBadRequest, 30001, " bad request")

	ErrModelIsExist = NewBcode(http.StatusConflict, 30002, "provider model already exist")

	ErrModelRecordNotFound = NewBcode(http.StatusNotFound, 30003, "model not exist")

	ErrAddModel = NewBcode(http.StatusInternalServerError, 30004, "model insert db failed")

	ErrDeleteModel = NewBcode(http.StatusInternalServerError, 30005, "model delete db failed")

	ErrEngineDeleteModel = NewBcode(http.StatusInternalServerError, 30006, "engine delete model failed")

	ErrNoRecommendModel = NewBcode(http.StatusNotFound, 30007, "No Recommend Model")

	ErrModelIsRunning = NewBcode(http.StatusNotFound, 30008, " model is still running, please wait before trying to perform the delete operation.")
)
