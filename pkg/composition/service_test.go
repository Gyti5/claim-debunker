package composition

import (
	"context"
	"errors"
	"testing"
)

type fakeChatGPTClient struct {
	items []string
	err   error
}

func (c *fakeChatGPTClient) ExtractIngredientsFromImage(context.Context, []byte, string) ([]string, error) {
	return c.items, c.err
}

func (c *fakeChatGPTClient) FindIngredientsByProductName(context.Context, string) ([]string, error) {
	return c.items, c.err
}

func TestServiceGetCompositionByText(t *testing.T) {
	t.Parallel()

	svc := NewService(&fakeChatGPTClient{
		items: []string{"Sugar", " Salt ", ""},
	})

	result, err := svc.GetComposition(context.Background(), Request{
		ProductName: "Coca-Cola",
	})
	if err != nil {
		t.Fatalf("GetComposition err: %v", err)
	}

	if len(result.Items) != 2 || result.Items[0] != "Sugar" || result.Items[1] != "Salt" {
		t.Fatalf("items=%v want=[Sugar Salt]", result.Items)
	}
}

func TestServiceGetCompositionByImage(t *testing.T) {
	t.Parallel()

	svc := NewService(&fakeChatGPTClient{
		items: []string{"Water"},
	})

	result, err := svc.GetComposition(context.Background(), Request{
		ImageBytes: []byte{1, 2, 3},
		ImageName:  "label.jpg",
	})
	if err != nil {
		t.Fatalf("GetComposition err: %v", err)
	}

	if len(result.Items) != 1 || result.Items[0] != "Water" {
		t.Fatalf("items=%v want=[Water]", result.Items)
	}
}

func TestServiceGetCompositionValidationErrors(t *testing.T) {
	t.Parallel()

	svc := NewService(&fakeChatGPTClient{})

	_, err := svc.GetComposition(context.Background(), Request{})
	if !errors.Is(err, ErrEmptyInput) {
		t.Fatalf("empty input err=%v want=%v", err, ErrEmptyInput)
	}

	_, err = svc.GetComposition(context.Background(), Request{
		ProductName: "X",
		ImageBytes:  []byte{1},
	})
	if !errors.Is(err, ErrConflictingData) {
		t.Fatalf("conflicting input err=%v want=%v", err, ErrConflictingData)
	}
}
