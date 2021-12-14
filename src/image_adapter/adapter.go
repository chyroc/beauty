package image_adapter

type Image struct {
	UserID  string `json:"user_id"`
	ImageID string `json:"image_id"`
	URL     string `json:"url"`
}

type GetImager interface {
	GetImage(data string) ([]*Image, error)
}
