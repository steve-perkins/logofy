package logofy

import (
	"fmt"
	"net/http"

	"appengine"
	"appengine/memcache"
)

func init() {
	http.Handle("/", http.FileServer(http.Dir("static")))
	http.HandleFunc("/logo", defaultHandler)
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	abstractHandler(w, r, "default.png")
}

func abstractHandler(w http.ResponseWriter, r *http.Request, logoFilename string) {
	// If the requested pic with the requested logo has already been generated, then use the cached version
	ctx := appengine.NewContext(r)
	originalImageUrl := r.URL.Query().Get("img")
	if item, err := memcache.Get(ctx, logoFilename+":"+originalImageUrl); err == nil {
		ctx.Infof("Retrieved [%s:%s] from memcache\n", logoFilename, originalImageUrl)
		w.Header().Set("Content-Type", "image/png")
		w.Write(item.Value)
	} else {
		// Load the logo image
		logoImage, err := fetchLogoImage(logoFilename)
		if err != nil {
			message := fmt.Sprintf("Unable to load logo image file: %s\n", err)
			ctx.Errorf(message)
			fmt.Fprintf(w, message)
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
				ctx.Infof("Caching and serving up [%s:%s] from memcache\n", logoFilename, originalImageUrl)
				w.Header().Set("Content-Type", "image/png")
				w.Write(generatedImageBytes)
			} else {
				message := fmt.Sprintf("An error occured: %s\n", err)
				ctx.Errorf(message)
				fmt.Fprintf(w, message)
			}
		} else {
			message := fmt.Sprintf("An error occurred: %s\n", err)
			ctx.Errorf(message)
			fmt.Fprintf(w, message)
		}
	}
}
