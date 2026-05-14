package composition

type Request struct {
	ProductName string
	ImageBytes  []byte
	ImageName   string
}

type Result struct {
	Items []string `json:"items"`
}
