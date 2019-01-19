package functions

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	_ "image/gif" // import only
	_ "image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/nlopes/slack"
	"github.com/pkg/errors"
)

type NotFound interface {
	Error() error
	IsNotFound() bool
}

type NotFoundError struct {
	err error
}

func (nfe *NotFoundError) Error() string {
	return nfe.err.Error()
}

func (nfe *NotFoundError) IsNotFound() bool {
	return true
}

// Amesh is endpoint of cloud functions
func Amesh(w http.ResponseWriter, r *http.Request) {
	s, err := slack.SlashCommandParse(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	now := time.Now().In(loc)
	img, err := getAmeshImage(now)
	if err != nil {
		log.Printf("%#v", err)
		if _, ok := err.(NotFound); ok {
		}
		return
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		log.Printf("%#v", err)
		return
	}
	err = uploadAmeshImage(&buf, s.ChannelID)
	if err != nil {
		log.Printf("line 50: %#v", err)
	}
}

func uploadAmeshImage(file io.Reader, channel string) error {
	token := os.Getenv("SLACK_BOT_TOKEN")
	api := slack.New(token)

	params := slack.FileUploadParameters{
		Filename: "amesh.png",
		Title:    "amesh",
		Reader:   file,
		Channels: []string{
			channel,
		},
	}
	f, err := api.UploadFile(params)
	if err != nil {
		log.Printf("%s\n", err)
		return err
	}
	fmt.Printf("Name: %s, URL: %s\n", f.Name, f.URL)
	return nil
}

func getAmeshImage(n time.Time) (image.Image, error) {
	baseURL := "http://tokyo-ame.jwa.or.jp/map/map000.jpg"
	baseResp, err := http.Get(baseURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get base image")
	}
	defer baseResp.Body.Close()
	base, _, err := image.Decode(baseResp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode base image")
	}

	ameURL := fmt.Sprintf("http://tokyo-ame.jwa.or.jp/mesh/000/%d%02d%02d%02d%02d.gif", n.Year(), n.Month(), n.Day(), n.Hour(), n.Minute()/5*5)
	ameResp, err := http.Get(ameURL)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get ame image %s", ameURL)
	}
	defer ameResp.Body.Close()
	if ameResp.StatusCode == http.StatusNotFound {
		return nil, &NotFoundError{err: errors.New(ameURL + " is not found")}
	}
	ame, _, err := image.Decode(ameResp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode ame image from %s", ameURL)
	}

	maskURL := "http://tokyo-ame.jwa.or.jp/map/msk000.png"
	maskResp, err := http.Get(maskURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get mask image")
	}
	defer maskResp.Body.Close()
	mask, _, err := image.Decode(maskResp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode mask image")
	}

	outRect := image.Rectangle{image.Pt(0, 0), ame.Bounds().Size()}
	out := image.NewRGBA(outRect)
	draw.Draw(out, base.Bounds(), base, image.Pt(0, 0), 0)
	draw.Draw(out, ame.Bounds(), ame, image.Pt(0, 0), 0)
	draw.Draw(out, mask.Bounds(), mask, image.Pt(0, 0), 0)

	return out, nil
}
