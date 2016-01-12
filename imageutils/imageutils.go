package imageutils

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
	"strings"
)

const TARGET_IMAGE_WIDTH uint = 800
const TARGET_LOGO_WIDTH uint = 400

// logoFilename should match the name of an image file within the "logos" subdirectory
func FetchLogoImage(logoFilename string) (image.Image, error) {
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

func FetchImage(ctx appengine.Context, imgUrl string, targetWidth uint) (image.Image, error) {
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
	img, _, err := bytesToImage(imgBytes) // imgConfig
	if err != nil {
		return nil, err
	}
	// Resize the image
	img = resize.Resize(targetWidth, 0, img, resize.Lanczos3)
	// Success!
	return img, nil
}



func GenerateImageWithLogo(originalImage image.Image, logoImage image.Image, pos string) image.Image {
	// Create an editable image from the original
	newImage := image.NewRGBA(originalImage.Bounds())
	draw.Draw(newImage, originalImage.Bounds(), originalImage, image.ZP, draw.Over)
	// Superimpose the logo image in the bottom-right corner of the new editable image
	var min, max image.Point

	switch strings.ToLower(pos) {
	case "":
		min = image.Point{
			X: originalImage.Bounds().Max.X - logoImage.Bounds().Max.X - 10,
			Y: originalImage.Bounds().Max.Y - logoImage.Bounds().Max.Y - 10,
		}
		max = image.Point {
			X: originalImage.Bounds().Max.X - 10,
			Y: originalImage.Bounds().Max.Y + 10,
		}
	case "br":
		min = image.Point{
			X: originalImage.Bounds().Max.X - logoImage.Bounds().Max.X - 10,
			Y: originalImage.Bounds().Max.Y - logoImage.Bounds().Max.Y - 10,
		}
		max = image.Point {
			X: originalImage.Bounds().Max.X - 10,
			Y: originalImage.Bounds().Max.Y + 10,
		}
	case "bl":
		min = image.Point{
			X: originalImage.Bounds().Min.X - logoImage.Bounds().Min.X + 10,
			Y: originalImage.Bounds().Max.Y - logoImage.Bounds().Max.Y - 10,
		}
		max = image.Point{
			X: originalImage.Bounds().Max.X + 10,
			Y: originalImage.Bounds().Max.Y + 10,
		}
	case "tl":
		min = image.Point{
			X: originalImage.Bounds().Min.X - logoImage.Bounds().Min.X + 10,
			Y: originalImage.Bounds().Min.Y - logoImage.Bounds().Min.Y + 10,
		}
		max = image.Point{
			X: originalImage.Bounds().Max.X + 10,
			Y: originalImage.Bounds().Max.Y - 10,
		}
	case "tr":
		min = image.Point{
			X: originalImage.Bounds().Max.X - logoImage.Bounds().Max.X - 10,
			Y: originalImage.Bounds().Min.Y - logoImage.Bounds().Min.Y + 10,
		}
		max = image.Point{
			X: originalImage.Bounds().Max.X + 10,
			Y: originalImage.Bounds().Max.Y - 10,
		}
	case "tc":
		min = image.Point{
			X: (originalImage.Bounds().Max.X - logoImage.Bounds().Max.X) / 2,
			Y: originalImage.Bounds().Min.Y - logoImage.Bounds().Min.Y + 10,
		}
		max = image.Point{
			X: originalImage.Bounds().Max.X + 10,
			Y: originalImage.Bounds().Max.Y - 10,
		}
	case "bc":
		min = image.Point{
			X: (originalImage.Bounds().Max.X - logoImage.Bounds().Max.X) / 2,
			Y: originalImage.Bounds().Max.Y - logoImage.Bounds().Max.Y - 10,
		}
		max=  image.Point{
			X: originalImage.Bounds().Max.X + 10,
			Y: originalImage.Bounds().Max.Y + 10,
		}
	}
	logoBounds := image.Rectangle{ Min: min, Max: max }
	draw.Draw(newImage, logoBounds, logoImage, image.ZP, draw.Over)
	return newImage
}
func ImageToBytes(img image.Image) ([]byte, error) {
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
