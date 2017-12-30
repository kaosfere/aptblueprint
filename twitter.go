package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/url"
	"strconv"

	"git.rcj.io/aptdata"
	"github.com/ChimeraCoder/anaconda"
	"github.com/davecgh/go-spew/spew"
)

type credentials struct {
	consumerKey       string
	consumerSecret    string
	accessToken       string
	accessTokenSecret string
}

func post(c credentials, apt *aptdata.Airport) error {
	anaconda.SetConsumerKey(c.consumerKey)
	anaconda.SetConsumerSecret(c.consumerSecret)
	api := anaconda.NewTwitterApi(c.accessToken, c.accessTokenSecret)

	data, err := ioutil.ReadFile("out.png")
	if err != nil {
		return err
	}

	mediaString := base64.StdEncoding.EncodeToString(data)

	media, error := api.UploadMedia(mediaString)
	if error != nil {
		return err
	}

	latitude := strconv.FormatFloat(apt.Latitude, 'f', -1, 64)
	longitude := strconv.FormatFloat(apt.Longitude, 'f', -1, 64)
	location := fmt.Sprintf("%s, %s", apt.Region, apt.Country)
	if apt.City != "" {
		location = fmt.Sprintf("%s, %s", apt.City, location)
	}

	v := url.Values{}
	v.Set("media_ids", media.MediaIDString)
	v.Set("lat", latitude)
	v.Set("long", longitude)
	v.Set("display_coordinates", "true")
	tweet, err := api.PostTweet("", v)
	spew.Dump(tweet)
	//_, err = api.PostTweet("", v)
	return err
}
