package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/DiLRandI/console-power-cut-tracker-lk/model"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/html"
)

func main() {
	area := flag.String("area", "T", "Area letter to get the details")
	_ = flag.Bool("all", false, "Display all areas")
	flag.Parse()

	form := url.Values{}
	form.Add("StartTime", time.Now().Format("2006-01-02"))
	form.Add("EndTime", time.Now().AddDate(0, 0, 1).Format("2006-01-02"))
	c := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},

			DisableKeepAlives: false,
		},
	}
	page, err := c.Get("https://cebcare.ceb.lk/Incognito/DemandMgmtSchedule")
	if err != nil {
		logrus.Fatal(err)
	}
	defer page.Body.Close()

	pageContent, err := io.ReadAll(page.Body)
	if err != nil {
		logrus.Fatal(err)
	}
	token, _ := getToken(pageContent)

	req, err := http.NewRequest(http.MethodPost, "https://cebcare.ceb.lk/Incognito/GetLoadSheddingEvents", strings.NewReader(form.Encode()))
	if err != nil {
		logrus.Fatal(err)
	}

	req.PostForm = form
	req.Header.Add("RequestVerificationToken", *token)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(page.Cookies()[0])

	res, err := c.Do(req)
	if err != nil {
		logrus.Fatal(err)
	}
	defer res.Body.Close()

	content, err := io.ReadAll(res.Body)
	if err != nil {
		logrus.Fatal(err)
	}

	var response []*model.Response
	if err := json.Unmarshal(content, &response); err != nil {
		logrus.Fatal(err)
	}

	filter := filterResponse(area, ptrB(false), response)

	out := ""
	for _, r := range filter {
		out = timeTill(r.StartTime, r.EndTime)
		if strings.ContainsAny(out, "Active") {
			break
		} else if strings.ContainsAny(out, "Passed") {
			continue
		} else {
			break
		}
	}

	fmt.Print(strings.TrimSpace(out))

}

func ptrB(b bool) *bool {
	return &b
}

func filterResponse(area *string, all *bool, response []*model.Response) []*model.Response {
	if area == nil || *all {
		return response
	}

	*area = strings.ToUpper(*area)
	filtered := []*model.Response{}
	for _, r := range response {
		if r.LoadShedGroupID == *area {
			filtered = append(filtered, r)
		}
	}

	return filtered
}

func timeTill(s, e string) string {
	st, err := getTime(s)
	if err != nil {
		logrus.Fatal(err)
	}

	et, err := getTime(e)
	if err != nil {
		logrus.Fatal(err)
	}

	now := time.Now()
	if et.Sub(now).Seconds() <= 0 {
		return "Passed"
	} else if st.Sub(now).Seconds() <= 0 {
		return "Active"
	} else if st.Sub(now).Seconds() <= 600 {
		return "|<" + getTimeDiffAsString(st, now)
	}

	return getTimeDiffAsString(st, now)
}

func getTimeDiffAsString(from, to time.Time) string {
	totalSecs := int64(from.Sub(to).Seconds())
	hours := totalSecs / 3600
	minutes := (totalSecs % 3600) / 60
	seconds := totalSecs % 60

	return fmt.Sprintf(" %02d : %02d : %02d", hours, minutes, seconds)
}

func getTime(s string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02T15:04:05", s, time.Local)
}

func getToken(pc []byte) (*string, error) {
	node, err := html.Parse(bytes.NewReader(pc))
	if err != nil {
		return nil, err
	}

	in, err := input(node)
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

func input(doc *html.Node) (*html.Node, error) {
	var input *html.Node
	var crawler func(*html.Node)
	crawler = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "input" {
			input = node
			return
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			crawler(child)
		}
	}
	crawler(doc)
	if input != nil {
		return input, nil
	}

	return nil, errors.New("missing <input> in the node tree")
}
