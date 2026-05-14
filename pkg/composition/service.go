package composition

import (
	"context"
	"strings"
)

type Service struct {
	client ChatGPTClient
}

func NewService(client ChatGPTClient) *Service {
	return &Service{client: client}
}

func (s *Service) GetComposition(ctx context.Context, req Request) (Result, error) {
	hasImage := len(req.ImageBytes) > 0
	hasText := strings.TrimSpace(req.ProductName) != ""

	if hasImage && hasText {
		return Result{}, ErrConflictingData
	}
	if !hasImage && !hasText {
		return Result{}, ErrEmptyInput
	}

	var (
		items []string
		err   error
	)
	if hasImage {
		items, err = s.client.ExtractIngredientsFromImage(ctx, req.ImageBytes, req.ImageName)
	} else {
		items, err = s.client.FindIngredientsByProductName(ctx, strings.TrimSpace(req.ProductName))
	}
	if err != nil {
		return Result{}, err
	}

	filtered := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			filtered = append(filtered, trimmed)
		}
	}

	if len(filtered) == 0 {
		return Result{}, ErrNoComposition
	}

	return Result{
		Items: filtered,
	}, nil
}
