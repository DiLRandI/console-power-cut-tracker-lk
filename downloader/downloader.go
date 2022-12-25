package downloader

import (
	"bytes"
	"crypto/tls"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/DiLRandI/console-power-cut-tracker-lk/crawler"
	"github.com/DiLRandI/console-power-cut-tracker-lk/util"
	"golang.org/x/net/html"
)

func DownloadPowerCutData(start, end string) []byte {
	form := url.Values{}
	form.Add("StartTime", start)
	form.Add("EndTime", end)
	c := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},

			DisableKeepAlives: false,
		},
	}
	page, err := c.Get("https://cebcare.ceb.lk/Incognito/DemandMgmtSchedule")
	util.FailOnError(err, "")
	defer page.Body.Close()

	pageContent, err := io.ReadAll(page.Body)
	util.FailOnError(err, "")
	token, _ := getToken(pageContent)

	req, err := http.NewRequest(http.MethodPost, "https://cebcare.ceb.lk/Incognito/GetLoadSheddingEvents", strings.NewReader(form.Encode()))
	util.FailOnError(err, "")

	req.PostForm = form
	req.Header.Add("RequestVerificationToken", *token)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(page.Cookies()[0])

	res, err := c.Do(req)
	util.FailOnError(err, "")
	defer res.Body.Close()

	content, err := io.ReadAll(res.Body)
	util.FailOnError(err, "")

	return content

}

func getToken(pc []byte) (*string, error) {
	node, err := html.Parse(bytes.NewReader(pc))
	if err != nil {
		return nil, err
	}

	in, err := crawler.GetFirstInput(node)
	if err != nil {
		return nil, err
	}
	for _, a := range in.Attr {
		if a.Key == "value" {
			return &a.Val, nil
		}
	}
	return nil, nil
}
