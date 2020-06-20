package main

import (
	"log"
)

func getFaces() ([]Face, error) {
	var faces []Face
	if err := DB.Find(&faces, &Face{}).Error; err != nil {
		log.Printf("find faces error: %v", DB.Error)
		return nil, err
	}
	return faces, nil
}

func getPeople() ([]Person, error) {
	var people []Person
	if err := DB.Find(&people, &Person{}).Error; err != nil {
		log.Printf("find person error: %v", DB.Error)
		return nil, err
	}
	return people, nil
}

func getImages() ([]Image, error) {
	var images []Image
	if err := DB.Find(&images, &Image{}).Error; err != nil {
		log.Printf("find faces error: %v", DB.Error)
		return nil, err
	}
	return images, nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
