package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/disintegration/imaging"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

func getImage(url string) (image []byte, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func rotateImage(imageBytes []byte) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		return nil, err
	}
	// rotate the image by -90 degrees counter clockwise
	// same as rotating the image by 90 degrees clockwise
	// but this imaging library's rotate function rotates ccw
	img = imaging.Rotate(img, -90, color.Black)
	// encode the image
	buf := new(bytes.Buffer)
	err = png.Encode(buf, img)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func init() {
	// loads .env file into env vars
	godotenv.Load()
}

func main() {
	api := slack.New(os.Getenv("SLACK_TOKEN"), slack.OptionDebug(true))
	for {
		// get the user's slack profile
		profile, err := api.GetUserProfile(&slack.GetUserProfileParameters{UserID: os.Getenv("SLACK_USER_ID")})
		if err != nil {
			logrus.WithError(err).Fatal("Failed to get user profile")
		}
		logrus.Infof("Got user profile: %s", profile.RealNameNormalized)
		// get the user's profile picture
		image, err := getImage(profile.ImageOriginal)
		if err != nil {
			logrus.WithError(err).Fatal("Failed to get user image")
		}
		// rotate the image
		image, err = rotateImage(image)
		if err != nil {
			logrus.WithError(err).Fatal("Failed to rotate image")
		}

		// save image as pfp.png
		err = ioutil.WriteFile("pfp.png", image, 0644)
		if err != nil {
			logrus.WithError(err).Fatal("Failed to save image")
		}

		// set the user's pfp to the image
		err = api.SetUserPhoto("pfp.png", slack.UserSetPhotoParams{}) // why oh why does this take a filename instead of a byte array
		if err != nil {
			logrus.WithError(err).Fatal("Failed to set user photo")
		}
		logrus.Info("Successfully set user photo")
		time.Sleep(time.Minute * 5) // horrible approach and i should replace this with a crontab or something but i can't be bothered
	}
}
