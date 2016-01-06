package logofy

import (
	"bytes"
	"errors"
	"image"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"os"
	"path"

	"appengine"
	"appengine/urlfetch"
	"fmt"
)

const MAX_IMAGE_BYTES = 1000000

// logoFilename should match the name of an image file within the "logos" subdirectory
func fetchLogoImage(logoFilename string) (image.Image, error) {
	if file, err := os.Open(path.Join("logos", logoFilename)); err == nil {
		defer file.Close()
		if img, _, err := image.Decode(file); err == nil { // _ == format
			return img, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func fetchOriginalImage(ctx appengine.Context, imgUrl string) (image.Image, error) {
	// Open HTTP request to image URL
	client := urlfetch.Client(ctx)
	if response, err := client.Get(imgUrl); err == nil {
		defer response.Body.Close()
		// Download image as stream of bytes, failing if the image size exceeds the maximum size
		imgBytes := make([]byte, MAX_IMAGE_BYTES)
		if numBytesRead, err := response.Body.Read(imgBytes); err == nil && numBytesRead < MAX_IMAGE_BYTES {
			// Decode the bytes into a image data type
			if img, imgConfig, err := bytesToImage(imgBytes); err == nil {
				// Verify that the dimensions of the image are large enough
				if imgConfig.Height < 200 && imgConfig.Width < 400 {
					err := errors.New("Images must be at least 500x500 in size")
					return nil, err
				}
				// Success!
				return img, nil
			} else {
				return nil, err
			}
		} else if numBytesRead >= MAX_IMAGE_BYTES {
			message := fmt.Sprintf("Could not download [%s] because it exceeds the maximum size\n", imgUrl)
			ctx.Errorf(message)
			return nil, errors.New(message)
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func generateImageWithLogo(originalImage image.Image, logoImage image.Image) image.Image {
	// Create an editable image from the original
	newImage := image.NewRGBA(originalImage.Bounds())
	draw.Draw(newImage, originalImage.Bounds(), originalImage, image.ZP, draw.Over)
	// Superimpose the logo image in the bottom-right corner of the new editable image
	logoBounds := image.Rectangle{
		Min: image.Point{
			X: originalImage.Bounds().Max.X - logoImage.Bounds().Max.X - 10,
			Y: originalImage.Bounds().Max.Y - logoImage.Bounds().Max.Y - 10,
		},
		Max: image.Point{
			X: originalImage.Bounds().Max.X - 10,
			Y: originalImage.Bounds().Max.Y - 10,
		},
	}
	draw.Draw(newImage, logoBounds, logoImage, image.ZP, draw.Over)
	return newImage
}

func imageToBytes(img image.Image) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, img); err == nil {
		return buf.Bytes(), nil
	} else {
		return nil, err
	}
}

func bytesToImage(imageBytes []byte) (image.Image, image.Config, error) {
	imageConfig, _, err := image.DecodeConfig(bytes.NewReader(imageBytes)) // _ == format
	if err != nil {
		return nil, image.Config{}, err
	}
	img, _, err := image.Decode(bytes.NewReader(imageBytes)) // _ == format
	if err != nil {
		return nil, image.Config{}, err
	}
	return img, imageConfig, nil
}
