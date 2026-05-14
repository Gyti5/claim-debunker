package composition

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestOpenAIClientRequiresAPIKey(t *testing.T) {
	t.Parallel()

	client := NewOpenAIChatGPTClient(OpenAIClientConfig{
		APIKey:  "",
		Model:   "gpt-5.5",
		BaseURL: "https://api.openai.com/v1/responses",
		Timeout: 5 * time.Second,
	})

	_, err := client.FindIngredientsByProductName(context.Background(), "Fanta")
	if !errors.Is(err, ErrChatGPTNotConfigured) {
		t.Fatalf("err=%v want=%v", err, ErrChatGPTNotConfigured)
	}
}

func TestParseIngredientsOutputArray(t *testing.T) {
	t.Parallel()

	items, err := parseIngredientsOutput(`["Water","Sugar"]`)
	if err != nil {
		t.Fatalf("parseIngredientsOutput err: %v", err)
	}
	if len(items) != 2 || items[0] != "Water" || items[1] != "Sugar" {
		t.Fatalf("items=%v want=[Water Sugar]", items)
	}
}

func TestParseIngredientsOutputWrappedJSON(t *testing.T) {
	t.Parallel()

	items, err := parseIngredientsOutput(`{"items":["Salt","Pepper"]}`)
	if err != nil {
		t.Fatalf("parseIngredientsOutput err: %v", err)
	}
	if len(items) != 2 || items[0] != "Salt" || items[1] != "Pepper" {
		t.Fatalf("items=%v want=[Salt Pepper]", items)
	}
}

func TestParseIngredientsOutputCodeFence(t *testing.T) {
	t.Parallel()

	items, err := parseIngredientsOutput("```json\n[\"A\",\"B\"]\n```")
	if err != nil {
		t.Fatalf("parseIngredientsOutput err: %v", err)
	}
	if len(items) != 2 || items[0] != "A" || items[1] != "B" {
		t.Fatalf("items=%v want=[A B]", items)
	}
}

func TestParseIngredientsOutputWrappedEmptyList(t *testing.T) {
	t.Parallel()

	items, err := parseIngredientsOutput(`{"items":[]}`)
	if err != nil {
		t.Fatalf("parseIngredientsOutput err: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("items=%v want empty list", items)
	}
}

func TestDetectImageMIMETypeRejectsNonImage(t *testing.T) {
	t.Parallel()

	_, err := detectImageMIMEType([]byte("not-an-image"), "sample.txt")
	if !errors.Is(err, ErrInvalidImageContentType) {
		t.Fatalf("err=%v want=%v", err, ErrInvalidImageContentType)
	}
}
