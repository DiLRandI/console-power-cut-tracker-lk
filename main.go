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
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"example.com/test/model"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/html"
)

func main() {
	area := flag.String("area", "T", "Area letter to get the details")
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

	printTable(filterResponse(area, response))

}

func filterResponse(area *string, response []*model.Response) []*model.Response {
	if area == nil {
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

func printTable(response []*model.Response) {
	tw := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	printHeaders(tw)
	for _, r := range response {
		fmt.Fprintf(tw, "| %s \t| %s \t| %s \t| %s \t| %s \t|\n",
			r.LoadShedGroupID,
			getTime(r.StartTime),
			getTime(r.EndTime),
			duration(r.StartTime, r.EndTime),
			timeTill(r.StartTime))
		printVerticalLine(tw)
	}
	tw.Flush()
}

func timeTill(s string) string {
	st, err := time.ParseInLocation("2006-01-02T15:04:05", s, time.Local)
	if err != nil {
		logrus.Fatal(err)
	}
	fmt.Println(st)
	now := time.Now()
	fmt.Println(now)
	totalSecs := int64(st.Sub(now).Seconds())
	hours := totalSecs / 3600
	minutes := (totalSecs % 3600) / 60
	seconds := totalSecs % 60

	return fmt.Sprintf(" %02d : %02d : %02d", hours, minutes, seconds)
}

func duration(s1, s2 string) string {
	return ""
}

func getTime(s string) string {
	time, err := time.Parse("2006-01-02T15:04:05", s)
	if err != nil {
		logrus.Fatal(err)
	}
	return time.Format("15:04:05")
}

func printHeaders(tw *tabwriter.Writer) {
	printVerticalLine(tw)
	fmt.Fprintf(tw, "| Group Letter \t| Start Time \t| End Time \t|  Duration \t|  Power Cut in \t|\n")
	printVerticalLine(tw)
}

func printVerticalLine(tw *tabwriter.Writer) {
	fmt.Fprintf(tw, "| ------------------ \t"+
		"| ------------------ \t"+
		"| ------------------ \t"+
		"| ------------------ \t"+
		"| ------------------ \t"+
		"|\n")
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
