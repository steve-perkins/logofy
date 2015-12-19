package logofy

import (
	"fmt"
	"log"
	"net/http"

	"appengine"
	"appengine/memcache"
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


