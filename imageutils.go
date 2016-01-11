package logofy

import (
	"bytes"
	"image"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"os"
	"path"

	"appengine"
	"appengine/urlfetch"
	"github.com/nfnt/resize"
	"io/ioutil"
)

const TARGET_IMAGE_WIDTH uint = 800
const TARGET_LOGO_WIDTH uint = 400

// logoFilename should match the name of an image file within the "logos" subdirectory
func fetchLogoImage(logoFilename string) (image.Image, error) {
	if file, err := os.Open(path.Join("logos", logoFilename)); err == nil {
		defer file.Close()
		if img, _, err := image.Decode(file); err == nil { // _ == format
			img = resize.Resize(TARGET_LOGO_WIDTH, 0, img, resize.Lanczos3)
			return img, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func fetchImage(ctx appengine.Context, imgUrl string, targetWidth uint) (image.Image, error) {
	// Open HTTP request to image URL
	client := urlfetch.Client(ctx)
	response, err := client.Get(imgUrl)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	// Download image as stream of bytes
	imgBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	// Decode the bytes into a image data type
	img, _, err := bytesToImage(imgBytes)  // imgConfig
	if err != nil {
		return nil, err
	}
	// Resize the image
	img = resize.Resize(targetWidth, 0, img, resize.Lanczos3)
	// Success!
	return img, nil
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
