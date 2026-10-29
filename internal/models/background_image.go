package models

type BackgroundImage struct {
	ContentType string `json:"content_type"`
	Data        []byte `json:"data" tstype:"Uint8Array<ArrayBuffer>"`
	UpdatedAt   int64  `json:"updated_at"`
}
