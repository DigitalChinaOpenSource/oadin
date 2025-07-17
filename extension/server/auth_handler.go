package server

import (
	"context"
	"encoding/json"
	"fmt"
)

type DefaultAuthHandler struct{}

type SmartVisionAuthHandler struct{}

func NewAuthHandler(flavor string) AuthHandlerProvider {
	switch flavor {
	case "smartvision":
		return &SmartVisionAuthHandler{}
	default:
		return &DefaultAuthHandler{}
	}
}

func (d *DefaultAuthHandler) MergeAuthKey(ctx context.Context, existingAuth, newAuth string) (string, error) {

	return newAuth, nil
}

func (d *DefaultAuthHandler) ValidateAuthKey(ctx context.Context, authKey string) error {
	if authKey == "" {
		return fmt.Errorf("auth key cannot be empty")
	}
	return nil
}

func (s *SmartVisionAuthHandler) MergeAuthKey(ctx context.Context, existingAuth, newAuth string) (string, error) {
	if existingAuth == "" {
		return newAuth, nil
	}
	var existingMap map[string]interface{}
	var newMap map[string]interface{}

	err := json.Unmarshal([]byte(existingAuth), &existingMap)
	if err != nil {
		return "", fmt.Errorf("failed to parse existing auth key: %w", err)
	}

	err = json.Unmarshal([]byte(newAuth), &newMap)
	if err != nil {
		return "", fmt.Errorf("new auth key must be valid JSON: %w", err)
	}

	for key, value := range newMap {
		existingMap[key] = value
	}

	mergedBytes, err := json.Marshal(existingMap)
	if err != nil {
		return "", fmt.Errorf("failed to marshal merged auth key: %w", err)
	}
	return string(mergedBytes), nil
}

func (s *SmartVisionAuthHandler) ValidateAuthKey(ctx context.Context, authKey string) error {
	if authKey == "" {
		return fmt.Errorf("smartvision auth key cannot be empty")
	}

	var authMap map[string]interface{}
	err := json.Unmarshal([]byte(authKey), &authMap)
	if err != nil {
		return fmt.Errorf("smartvision auth key must be valid JSON: %w", err)
	}

	return nil
}
