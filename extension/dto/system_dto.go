package dto

type ProxyRequest struct {
	Endpoint string `json:"endpoint" validate:"required"`
	Username string `json:"username" `
	Password string `json:"password" `
}

type FeedbackRequest struct {
	Feedback string `json:"feedback" validate:"required"`
}