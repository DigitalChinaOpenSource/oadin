package bcode

import "net/http"

var (
	ControlPanelCode = NewBcode(http.StatusOK, 40000, "control panel interface call success")

	ControlPanelPathCheckError = NewBcode(http.StatusBadRequest, 40001, "path is not correct or not exists, please check!")

	ControlPanelPathSizeError = NewBcode(http.StatusBadRequest, 40002, "path size is not enough, please modify!")

	ControlPanelPathStatusError = NewBcode(http.StatusBadRequest, 40003, "path is not empty or same as the source path, please check!")

	ControlPanelCopyDirError = NewBcode(http.StatusOK, 40004, "migrate file failed, please retry!")
)
