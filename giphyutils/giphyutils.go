package giphyutils

import (
	"appengine"
	"appengine/urlfetch"
	"encoding/json"
	"fmt"
	"image"
	"imageutils"
	"io/ioutil"
	"net/url"
	"strconv"
)

type GiphyPayload struct {
	Data []GiphyData `json:"data"`
}

type GiphyData struct {
	Images GiphyImages `json:"images"`
}

type GiphyImages struct {
	OriginalStill GiphyOriginalStill `json:"original_still"`
}

type GiphyOriginalStill struct {
	URL    string `json:"url"`
	Width  string `json:"width"`
	Height string `json:"height"`
}

func FetchFromGify(ctx appengine.Context, keyword string) (image.Image, error) {
	client := urlfetch.Client(ctx)
	gifyUrl := fmt.Sprintf("http://api.giphy.com/v1/gifs/search?q=%s&limit=1&api_key=dc6zaTOxFJmzC", url.QueryEscape(keyword))
	ctx.Infof("Hitting Gify URL: %s\n", gifyUrl)
	response, err := client.Get(gifyUrl)
	if err != nil {
		ctx.Errorf("An error occurred: %s\n", err)
		return nil, err
	}
	defer response.Body.Close()
	// Download image as stream of bytes
	jsonBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		ctx.Errorf("An error occurred: %s\n", err)
		return nil, err
	}
	//	jsonString := string(jsonBytes)
	//	ctx.Infof("Retrieved JSON from Gify: %s\n", jsonString)

	var payload GiphyPayload
	err = json.Unmarshal(jsonBytes, &payload)
	if err != nil {
		ctx.Errorf("An error occurred: %s\n", err)
		return nil, err
	}
	originalUrl := payload.Data[0].Images.OriginalStill.URL
	originalWidth, err := strconv.Atoi(payload.Data[0].Images.OriginalStill.Width)
	if err != nil {
		ctx.Errorf("An error occurred: %s\n", err)
		return nil, err
	}
	ctx.Infof("Retrieved original image URL: %s, with width: %d\n", originalUrl, originalWidth)

	return imageutils.FetchImage(ctx, originalUrl, uint(originalWidth))
}
