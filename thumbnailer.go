package main

import (
	"fmt"

	"github.com/h2non/bimg"
)

type ExtractRect struct {
	top, left, width, height int
}

func Thumbnail(image []byte, rect ExtractRect) ([]byte, error) {
	baseImg, err := bimg.NewImage(image).Extract(rect.top, rect.left, rect.width, rect.height)
	if err != nil {
		return nil, fmt.Errorf("extract baseImg: %w", err)
	}

	fg := bimg.NewImage(baseImg)

	size, err := fg.Size()
	if err != nil {
		return nil, fmt.Errorf("get size of foreground: %w", err)
	}

	width := float64(size.Width)
	height := float64(size.Height)
	top := 0.
	left := 0.

	if width > height {
		height = height * (400 / width)
		width = 400
		top = (300 - height) / 2
	} else {
		width = width * (300 / height)
		height = 300
		left = (400 - width) / 2
	}

	wm, err := fg.Process(bimg.Options{
		Width:   int(width),
		Height:  int(height),
		Enlarge: true,
	})
	if err != nil {
		return nil, fmt.Errorf("create watermark: %w", err)
	}

	bg := bimg.NewImage(baseImg)

	finalImg, err := bg.Process(bimg.Options{
		Width:        400,
		Height:       300,
		Gravity:      bimg.GravityCentre,
		GaussianBlur: bimg.GaussianBlur{Sigma: 15},
		Crop:         true,
		Quality:      95,
		Enlarge:      true,
		WatermarkImage: bimg.WatermarkImage{
			Left: int(left),
			Top:  int(top),
			Buf:  wm,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create final image: %w", err)
	}

	return finalImg, nil
}
