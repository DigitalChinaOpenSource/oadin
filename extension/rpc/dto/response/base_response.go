package response

type BaseResponse[T any] struct {
	Code int `json:"code"` // 状态码
	Data T   `json:"data"` // 数据
}

type ChangeList struct{}

type ChangeListResponse1 = BaseResponse[[]ChangeList]

func sdf() {
	var res ChangeListResponse1
	res.Code = 200
	res.Data = []ChangeList{}
	// 处理 res ...
	// 例如：打印状态码和数据长度
	println(res.Code, len(res.Data))
}
