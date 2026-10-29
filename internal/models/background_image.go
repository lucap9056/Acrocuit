package models

type BackgroundImage struct {
	ContentType string `json:"content_type"`
	Data        []byte `json:"data"`
	UpdatedAt   int64  `json:"updated_at"`
}
