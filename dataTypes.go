package main

type Face struct {
	ID             string `json:"faceId"`
	ParentPersonID string `json:"personId"`
	ParentImageID  uint   `json:"imageId"`
	FaceHeight     uint   `json:"faceHeight,omitempty"`
	FaceWidth      uint   `json:"faceWidth,omitempty"`
	FaceX          uint   `json:"faceX,omitempty"`
	FaceY          uint   `json:"faceY,omitempty"`
	Filename       string `json:"-"`
}

type Person struct {
	ID          string `json:"personId"`
	Featured    string `json:"featured,omitempty"`
	ChineseName string `json:"chineseName,omitempty"`
	EnglishName string `json:"englishName,omitempty"`
	IsTeacher   bool   `json:"isTeacher,omitempty"`
	Position    string `json:"position,omitempty"`
}

type Image struct {
	ID          uint   `json:"imageId"`
	CdnLocation string `json:"src"`
	Height      uint   `json:"height"`
	Width       uint   `json:"width"`
	Time        string `json:"time,omitempty"`
	Latitude    string `json:"latitude,omitempty"`
	Longitude   string `json:"longitude,omitempty"`
	Source      string `json:"source"`
	InSchool    *bool  `json:"in_school,omitempty"`
}
