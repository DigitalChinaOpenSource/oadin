package bcode

import "net/http"

var ControlPanelCode = NewBcode(http.StatusOK, 40000, "control panel interface call success")
