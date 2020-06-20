package main

type PaginatorRequest struct {
	Page  int `query:"page" validate:"required,min=1"`
	Limit int `query:"limit" validate:"required,min=1,max=100"`

	SortOrder string `query:"sort_order" validate:"oneof=desc asc"`
}

type UploadRequest struct {
	Filename string `json:"filename" validate:"required"`
	PersonID string `json:"personId" validate:"required,uuid"`
}

type UploadCallbackRequest struct {
	Code        string `form:"code" validate:"required"`
	URL         string `form:"url" validate:"required"`
	Time        int64  `form:"time" validate:"required"`
	ImageWidth  uint   `form:"image-width" validate:"required"`
	ImageHeight uint   `form:"image-height" validate:"required"`
	PersonID    string `form:"ext-param" validate:"required"`
}
