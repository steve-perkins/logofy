package logofy

import (
	"appengine"
	"appengine/memcache"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
 	"strings"
)

func init() {
	http.Handle("/", http.FileServer(http.Dir("static")))
	http.HandleFunc("/logo", defaultHandler)
	http.HandleFunc("/slack", slackHandler)
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	abstractHandler(w, r, "default.png")
}

func slackHandler(w http.ResponseWriter, r *http.Request) {
	titles := []string{
		"Another paradigm shift from Logofy",
		"This game-changer brought to you by Logofy",
		"Building synergy with Logofy",
		"Thinking outside the box with Logofy",
		"Pivoting this image's strategy with Logofy",
		"Adding business value with Logofy",
		"Moving the needle with Logofy",
		"Empowering innovative transformation with Logofy",
		"Gamifying the tipping point with Logofy",
		"This market disruption brought to you by Logofy",
	}
	randomIndex := rand.Intn(len(titles))

	paramString := strings.SplitAfter(url.QueryEscape(r.URL.Query().Get("text")), " ")
	originalImage, logoImage := paramString[0], paramString[1]
	imgUrl := `http://logofy-web.appspot.com/logo?img=` + originalImage + `&logo=` + logoImage
	jsonString :=
		`{
	"response_type" : "in_channel",
	"attachments" : [{
		"title" : "` + titles[randomIndex] + `",
		"image_url" : "` + imgUrl + `"
	}]
}`
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, jsonString)
}

func abstractHandler(w http.ResponseWriter, r *http.Request, logoFilename string) {
	// If the requested pic with the requested logo has already been generated, then use the cached version
	ctx := appengine.NewContext(r)
	originalImageUrl := r.URL.Query().Get("img")
	logoImageUrl := r.URL.Query().Get("logo")

	if item, err := memcache.Get(ctx, logoFilename+":"+originalImageUrl); err == nil {
		ctx.Infof("Retrieved [%s:%s] from memcache\n", logoFilename, originalImageUrl)
		w.Header().Set("Content-Type", "image/png")
		w.Write(item.Value)
	} else {
		// Load the logo image
		logoImage, err := fetchLogoImage(logoFilename)
		if logoImageUrl != "" {
			logoImage, err = fetchImage(ctx, logoImageUrl, 250)
		}
		if err != nil {
			message := fmt.Sprintf("Unable to load logo image file: %s\n", err)
			ctx.Errorf(message)
			fmt.Fprintf(w, message)
		}
		// Fetch the source image
		if originalImage, err := fetchImage(ctx, originalImageUrl, 800); err == nil {
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
