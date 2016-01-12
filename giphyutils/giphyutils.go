package giphyutils

import (
	"appengine"
	"appengine/urlfetch"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"
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

func FetchGiphyUrl(ctx appengine.Context, searchString string) (string, error) {
	client := urlfetch.Client(ctx)
	gifyUrl := fmt.Sprintf("http://api.giphy.com/v1/gifs/search?q=%s&limit=1&api_key=dc6zaTOxFJmzC", url.QueryEscape(searchString))
	ctx.Infof("Querying Gify URL: %s\n", gifyUrl)
	response, err := client.Get(gifyUrl)
	if err != nil {
		ctx.Errorf("An error occurred: %s\n", err)
		return "", err
	}
	defer response.Body.Close()
	// Download image as stream of bytes
	jsonBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		ctx.Errorf("An error occurred: %s\n", err)
		return "", err
	}
	//	jsonString := string(jsonBytes)
	//	ctx.Infof("Retrieved JSON from Gify: %s\n", jsonString)

	var payload GiphyPayload
	err = json.Unmarshal(jsonBytes, &payload)
	if err != nil {
		ctx.Errorf("An error occurred: %s\n", err)
		return "", err
	}
	if len(payload.Data) > 0 && strings.TrimSpace(payload.Data[0].Images.OriginalStill.URL) != "" {
		originalUrl := payload.Data[0].Images.OriginalStill.URL
		return originalUrl, nil
	} else {
		return "", errors.New("Giphy found no images for this query")
	}

	//	originalWidth, err := strconv.Atoi(payload.Data[0].Images.OriginalStill.Width)
	//	if err != nil {
	//		ctx.Errorf("An error occurred: %s\n", err)
	//		return nil, err
	//	}
	//	ctx.Infof("Retrieved original image URL: %s, with width: %d\n", originalUrl, originalWidth)
	//
	//	return imageutils.FetchImage(ctx, originalUrl, uint(originalWidth))
}
