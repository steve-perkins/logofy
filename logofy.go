package logofy

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"

	"appengine"
	"appengine/memcache"
	"appengine/urlfetch"
)

func init() {
	http.Handle("/", http.FileServer(http.Dir("static")))
	http.HandleFunc("/brazzers", brazzersHandler)
}

func brazzersHandler(w http.ResponseWriter, r *http.Request) {
	abstractHandler(w, r, "brazzers.png")
}

func abstractHandler(w http.ResponseWriter, r *http.Request, logoFilename string) {
	// If the requested pic with the requested logo has already been generated, then use the cached version
	originalImageUrl := r.URL.Query().Get("img")
	ctx := appengine.NewContext(r)
	if item, err := memcache.Get(ctx, logoFilename + ":" + originalImageUrl); err == nil {
		log.Printf("Retrieved [%s:%s] from memcache\n", logoFilename, originalImageUrl)
		w.Header().Set("Content-Type", "image/png")
		w.Write(item.Value)
	} else {
		// Load the logo image
		logoImage, err := fetchLogoImage(logoFilename)
		if err != nil {
			log.Fatalf("Unable to load logo image file: %s\n", err)
		}
		// Fetch the source image
		if originalImage, err := fetchOriginalImage(ctx, originalImageUrl); err == nil {
			// Generate and return an image with logo over the source
			generatedImage := generateImageWithLogo(originalImage, logoImage)
			if generatedImageBytes, err := imageToBytes(generatedImage); err == nil {
				// Cache the generated image bytes before sending them in the HTTP response
				item := &memcache.Item{
					Key:   logoFilename + ":" + originalImageUrl,
					Value: generatedImageBytes,
				}
				memcache.Add(ctx, item)
				log.Printf("Caching and serving up [%s:%s] from memcache\n", logoFilename, originalImageUrl)
				w.Header().Set("Content-Type", "image/png")
				w.Write(generatedImageBytes)
			} else {
				fmt.Fprintf(w, "An error occured: %s\n", err)
			}
		} else {
			fmt.Fprintf(w, "An error occured: %s\n", err)
		}
	}
}

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
		// Download image as stream of bytes
		if imgBytes, err := ioutil.ReadAll(response.Body); err == nil {
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
