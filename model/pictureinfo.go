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
