package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/DiLRandI/console-power-cut-tracker-lk/downloader"
	"github.com/DiLRandI/console-power-cut-tracker-lk/model"
	"github.com/DiLRandI/console-power-cut-tracker-lk/util"
)

func main() {
	area := flag.String("area", "T", "Area letter to get the details")
	_ = flag.Bool("all", false, "Display all areas")
	flag.Parse()

	content := downloader.DownloadPowerCutData(
		time.Now().Format("2006-01-02"),
		time.Now().AddDate(0, 0, 1).Format("2006-01-02"))

	var response []*model.Response
	util.FailOnError(json.Unmarshal(content, &response), "")

	if len(response) == 0 {
		fmt.Println("No PC today")
		os.Exit(0)
	}

	filter := filterResponse(area, util.PtrB(false), response)

	out := ""
	for _, r := range filter {
		out = timeTill(r.StartTime, r.EndTime)
		if strings.ContainsAny(out, "PC is Active") {
			break
		} else if strings.ContainsAny(out, "No more PC today!") {
			continue
		} else {
			break
		}
	}

	fmt.Print(strings.TrimSpace(out))

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
	st, err := util.ParseToLocalTime(s)
	util.FailOnError(err, "")

	et, err := util.ParseToLocalTime(e)
	util.FailOnError(err, "")

	now := time.Now()
	if et.Sub(now).Seconds() <= 0 {
		return "No more PC today!"
	} else if st.Sub(now).Seconds() <= 0 {
		return "PC is Active, restore in " + util.GetTimeDiffAsString(et, now)
	} else if st.Sub(now).Seconds() <= 600 {
		return "PC In less than 10 mins " + util.GetTimeDiffAsString(st, now)
	}

	return "PC In " + util.GetTimeDiffAsString(st, now)
}
