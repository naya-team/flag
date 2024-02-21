package server

type CreateFlag struct {
	Flag     string `json:"flag,omitempty"`
	IsEnable bool   `json:"is_enable,omitempty"`
}
