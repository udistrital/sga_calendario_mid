package models

type RelacionID struct {
	Id int `json:"Id"`
}

type APIErrorMarker struct {
	Type string `json:"Type,omitempty"`
}

type GenericPayload struct {
	TerceroId int `json:"TerceroId"`
}
