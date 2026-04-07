package model

type PictureInfo struct {
	Id             string `json:"id"`
	ImageSizeBytes int    `json:"imageSizeBytes"`
}

type PictureInfoLocal struct {
	Id             string `json:"id"`
	ImageSizeBytes int    `json:"imageSizeBytes"`
	TursoId        string `json:"tursoid"`
}

type Picture struct {
	Id         string `json:"id"`
	ImageBytes []byte `json:"image"`
	Created    string `json:"createdid"`
	Updated    string `json:"updatedid"`
}
