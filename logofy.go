package logofy

import (
	"appengine"
	"appengine/memcache"
	"fmt"
	"giphyutils"
	"imageutils"
	"math/rand"
	"net/http"
	"strings"
	"net/url"
)

const MEMCACHE_ENABLED = false

func init() {
	http.Handle("/", http.FileServer(http.Dir("static")))
	http.HandleFunc("/logo", defaultHandler)
	http.HandleFunc("/slack", slackHandler)
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	abstractHandler(w, r)
}

func slackHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
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

	textParam := r.URL.Query().Get("text")
	paramStrings := strings.SplitAfter(textParam, " ")

	var imgUrl string
	if len(paramStrings) > 2 {
		originalImage, logoImage, pos := strings.TrimSpace(paramStrings[0]), strings.TrimSpace(paramStrings[1]), strings.TrimSpace(paramStrings[2])
		imgUrl = `http://logofy-web.appspot.com/logo?img=` + originalImage + `&logo=` + logoImage + `&pos=` + pos
		ctx.Infof("originalImage: %s, logoImage: %s, pos:%s\n, ", originalImage, logoImage, pos)
	} else if len(paramStrings) > 1 {
		originalImage, logoOrPos := strings.TrimSpace(paramStrings[0]), strings.TrimSpace(paramStrings[1])
		logoOrPos = strings.ToLower(logoOrPos)
		if(logoOrPos == "tl" || logoOrPos == "tc" || logoOrPos == "tr" || logoOrPos == "bl" || logoOrPos == "bc" || logoOrPos == "br"){
			imgUrl = `http://logofy-web.appspot.com/logo?img=` + originalImage + `&pos=` + logoOrPos
			ctx.Infof("originalImage: %s, pos:%s\n, ", originalImage, logoOrPos)
		}else{
			imgUrl = `http://logofy-web.appspot.com/logo?img=` + originalImage + `&logo=` + logoOrPos
			ctx.Infof("originalImage: %s, logo:%s\n, ", originalImage, logoOrPos)
		}
	} else {
		imgUrl = `http://logofy-web.appspot.com/logo?img=` + url.QueryEscape(textParam)
	}

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

func abstractHandler(w http.ResponseWriter, r *http.Request) {
	// If the requested pic with the requested logo has already been generated, then use the cached version
	ctx := appengine.NewContext(r)
	originalImageUrl := r.URL.Query().Get("img")

	logoImageUrl := r.URL.Query().Get("logo")
	if logoImageUrl == "" {
		logoImageUrl = "brazz"
	}

	if MEMCACHE_ENABLED {
		item, err := memcache.Get(ctx, logoImageUrl+":"+originalImageUrl)
		if err != nil {
			ctx.Infof("Retrieved [%s:%s] from memcache\n", logoImageUrl, originalImageUrl)
			w.Header().Set("Content-Type", "image/png")
			w.Write(item.Value)
			return
		}
	}
	// Load the logo image
	logoImage, err := imageutils.FetchLogoImage(logoImageUrl)
	if err != nil {
		logoImage, err = imageutils.FetchImage(ctx, logoImageUrl, imageutils.TARGET_LOGO_WIDTH)
	}
	if err != nil {
		message := fmt.Sprintf("Unable to load logo image file: %s\n", err)
		ctx.Errorf(message)
		fmt.Fprintf(w, message)
		return
	}

	giphyImageUrl, err := giphyutils.FetchGiphyUrl(ctx, originalImageUrl)
	if err != nil {
		ctx.Errorf("An error was encountered: %s\n", err)
	} else {
		ctx.Infof("Giphy returned this image URL: %s\n", giphyImageUrl)
		originalImageUrl = giphyImageUrl
	}

	originalImage, err := imageutils.FetchImage(ctx, originalImageUrl, imageutils.TARGET_IMAGE_WIDTH)
	if err != nil {
		message := fmt.Sprintf("An error occurred: %s\n", err)
		ctx.Errorf(message)
		fmt.Fprintf(w, message)
		return
	}

	// Set the logo position
	pos := r.URL.Query().Get("pos")

	// Generate and return an image with logo over the source
	generatedImage := imageutils.GenerateImageWithLogo(originalImage, logoImage, pos)
	generatedImageBytes, err := imageutils.ImageToBytes(generatedImage)
	if err != nil {
		message := fmt.Sprintf("An error occured: %s\n", err)
		ctx.Errorf(message)
		fmt.Fprintf(w, message)
		return
	}
	// Cache the generated image bytes before sending them in the HTTP response
	if MEMCACHE_ENABLED {
		item := &memcache.Item{
			Key:   logoImageUrl + ":" + originalImageUrl,
			Value: generatedImageBytes,
		}
		memcache.Add(ctx, item)
		ctx.Infof("Caching [%s:%s]\n", logoImageUrl, originalImageUrl)
	}
	ctx.Infof("Serving up [%s:%s]\n", logoImageUrl, originalImageUrl)
	w.Header().Set("Content-Type", "image/png")
	w.Write(generatedImageBytes)
}
