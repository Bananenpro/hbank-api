package services

import (
	"bytes"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/disintegration/imaging"
)

type Picture struct {
	Tiny   []byte
	Small  []byte
	Medium []byte
	Large  []byte
	Huge   []byte
}

type PictureSize string

const (
	PictureTiny   PictureSize = "tiny"
	PictureSmall  PictureSize = "small"
	PictureMedium PictureSize = "medium"
	PictureLarge  PictureSize = "large"
	PictureHuge   PictureSize = "huge"
)

func (ps PictureSize) Validate() bool {
	return ps == PictureTiny || ps == PictureSmall || ps == PictureMedium || ps == PictureLarge || ps == PictureHuge
}

func SupportedPictureMimeType(mimeType string) bool {
	return mimeType == "image/jpeg" || mimeType == "image/png" || mimeType == "image/gif"
}

func NewPicture(data []byte, mimeType string) (*Picture, error) {
	var img image.Image
	var err error

	if mimeType == "image/jpeg" || mimeType == "image/png" || mimeType == "image/gif" {
		img, err = loadStdImage(data)
		if err != nil {
			return nil, err
		}
	}

	if img.Bounds().Dx() > img.Bounds().Dy() {
		img = imaging.CropAnchor(img, img.Bounds().Dy(), img.Bounds().Dy(), imaging.Center)
	} else {
		img = imaging.CropAnchor(img, img.Bounds().Dx(), img.Bounds().Dx(), imaging.Center)
	}

	return createPictureFromImage(img)
}

func createPictureFromImage(img image.Image) (*Picture, error) {
	var picture Picture

	huge := bytes.Buffer{}
	err := imaging.Encode(&huge, imaging.Resize(img, 1024, 1024, imaging.Linear), imaging.JPEG)
	if err != nil {
		return nil, err
	}
	picture.Huge = huge.Bytes()

	large := bytes.Buffer{}
	err = imaging.Encode(&large, imaging.Resize(img, 512, 512, imaging.Linear), imaging.JPEG)
	if err != nil {
		return nil, err
	}
	picture.Large = large.Bytes()

	medium := bytes.Buffer{}
	err = imaging.Encode(&medium, imaging.Resize(img, 256, 256, imaging.Linear), imaging.JPEG)
	if err != nil {
		return nil, err
	}
	picture.Medium = medium.Bytes()

	small := bytes.Buffer{}
	err = imaging.Encode(&small, imaging.Resize(img, 128, 128, imaging.Linear), imaging.JPEG)
	if err != nil {
		return nil, err
	}
	picture.Small = small.Bytes()

	tiny := bytes.Buffer{}
	err = imaging.Encode(&tiny, imaging.Resize(img, 64, 64, imaging.Linear), imaging.JPEG)
	if err != nil {
		return nil, err
	}
	picture.Tiny = tiny.Bytes()

	return &picture, nil
}

func loadStdImage(data []byte) (image.Image, error) {
	return imaging.Decode(bytes.NewReader(data), imaging.AutoOrientation(true))
}
