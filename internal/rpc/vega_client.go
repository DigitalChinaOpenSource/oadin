package rpc

import (
	"oadin/internal/rpc/dto/response"
)

// About  retrieves the "About Us" information from the Vega API.
func About() (response.AboutUsResponse, error) {
	var res response.AboutUsResponse
	_, err := GlobalClient.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&res).
		Get("/api/system/about")

	if err != nil {
		return response.AboutUsResponse{}, err
	}

	return res, nil
}

func ChangeList() (response.ChangeListResponse, error) {
	var res response.ChangeListResponse
	_, err := GlobalClient.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&res).
		Get("/api/system/changelist")

	if err != nil {
		return response.ChangeListResponse{}, err
	}
	return response.ChangeListResponse{}, err

}
