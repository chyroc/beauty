package image_adapter

type Image struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
}

type GetImager interface {
	GetImage(data string) ([]*Image, error)
}
