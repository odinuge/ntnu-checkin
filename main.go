package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"golang.org/x/net/publicsuffix"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func main() {
	username := os.Getenv("FEIDE_USERNAME")
	password := os.Getenv("FEIDE_PASSWORD")
	if len(username) == 0 || len(password) == 0 {
		fmt.Printf(`Usage:
  Env variables:
  FEIDE_USERNAME: feide username for logging in
  FEIDE_PASSWORD: feide password for logging in
`)
		os.Exit(1)
	}
	help := func() {
		fmt.Printf(`Usage of %s:
  %s checkin --room=<room-id> --from=07:00 --to=23:00
    to checkin
  %s search [query]
    to search for rooms
  %s get
    to list checkins
  %s list
    to list checkins
  %s delete [chckin-id]
    to list checkins
`, os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0])
		os.Exit(1)
	}

	if len(os.Args) == 1 {
		help()
	}
	u, err := url.Parse("https://innsida.ntnu.no/checkin")
	if err != nil {
		log.Fatal(err)
	}
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{
		Jar: jar,
	}

	var resp *http.Response
	if resp, err = client.Get(u.String()); err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	str, err := ioutil.ReadAll(resp.Body)

	re := regexp.MustCompile(`section-list" href="([^"]*)"`)
	urll := string(re.FindSubmatch(str)[1]) + "feide%7Crealm%7Cntnu.no&remember=on&response_type=code"

	var resp2 *http.Response
	if resp2, err = client.Get(urll); err != nil {
		log.Fatal(err)
	}
	defer resp2.Body.Close()

	postingUrl := resp2.Request.URL.String()

	values := url.Values{
		"has_js":        []string{"0"},
		"inside_iframe": []string{"0"},
		"feidename":     []string{username},
		"password":      []string{password},
	}
	var resp3 *http.Response
	if resp3, err = client.PostForm(postingUrl, values); err != nil {
		log.Fatal(err)
	}
	defer resp3.Body.Close()
	str3, err := ioutil.ReadAll(resp3.Body)

	relayRe := regexp.MustCompile(`name="RelayState" value="([^"]*)"`)
	samlRe := regexp.MustCompile(`name="SAMLResponse" value="([^"]*)"`)
	actionRe := regexp.MustCompile(`action="([^"]*)"`)

	postingValues := url.Values{
		"SAMLResponse": []string{string(samlRe.FindSubmatch(str3)[1])},
		"RelayState":   []string{string(relayRe.FindSubmatch(str3)[1])},
	}

	var resp4 *http.Response
	if resp4, err = client.PostForm(string(actionRe.FindSubmatch(str3)[1]), postingValues); err != nil {
		log.Fatal(err)
	}
	defer resp4.Body.Close()
	str4, err := ioutil.ReadAll(resp4.Body)

	bar := func(name string) []string {
		re := regexp.MustCompile(fmt.Sprintf(`name="%s" value="([^"]*)"`, name))
		return []string{string(re.FindSubmatch(str4)[1])}
	}

	end := "https://auth.dataporten.no/oauth/authorization" + "?" + url.Values{
		"response_type":    bar("amp;response_type"),
		"redirect_uri":     bar("amp;redirect_uri"),
		"scope":            bar("amp;scope"),
		"state":            bar("amp;state"),
		"authselection":    bar("authselection"),
		"client_id":        bar("client_id"),
		"preselected":      bar("preselected"),
		"reauthentication": bar("reauthentication"),
	}.Encode()

	var resp8 *http.Response
	if resp8, err = client.Get(end); err != nil {
		log.Fatal(err)
	}
	defer resp8.Body.Close()
	switch os.Args[1] {
	case "checkin":
		mySet := flag.NewFlagSet("", flag.ExitOnError)
		var roomFlag = mySet.String("room", "", "Room ID to checkin (example 1234)")
		var startTimestamp = mySet.String("from", "", `Start timestamp (example "07:00" or "2020-10-30T11:21:55.423451+02:00")`)
		var endTimestamp = mySet.String("to", "", `End timestamp (example "23:00" or "2020-10-30T11:21:55.423451+02:00"`)

		mySet.Parse(os.Args[2:])

		if *roomFlag == "" || *endTimestamp == "" || *startTimestamp == "" {
			fmt.Printf("Usage of %s %s:\n", os.Args[0], os.Args[1])
			mySet.PrintDefaults()
			os.Exit(1)
		}

		now := time.Now()

		from, err := time.Parse(time.RFC3339, *startTimestamp)

		if err != nil {
			startSplitted := strings.Split(*startTimestamp, ":")
			startHour, err2 := strconv.Atoi(startSplitted[0])
			startMinute, err3 := strconv.Atoi(startSplitted[1])
			from = time.Date(now.Year(), now.Month(), now.Day(), startHour, startMinute, 0, 0, now.Location())

			if err2 != nil || err3 != nil {
				log.Fatalf("Unable to parse timestamp %q: %e", *startTimestamp, err)
			}
		}
		to, err := time.Parse(time.RFC3339, *endTimestamp)

		if err != nil {
			toSplitted := strings.Split(*endTimestamp, ":")
			endHour, err2 := strconv.Atoi(toSplitted[0])
			endMinute, err3 := strconv.Atoi(toSplitted[1])
			to = time.Date(now.Year(), now.Month(), now.Day(), endHour, endMinute, 0, 0, now.Location())

			if err2 != nil || err3 != nil {
				log.Fatalf("Unable to parse timestamp %q: %e", *startTimestamp, err)
			}
		}

		var resp *http.Response
		if resp, err = client.Get("https://innsida.ntnu.no/checkin/api/room/" + url.QueryEscape(*roomFlag)); err != nil {
			log.Fatal(err)
		}

		var roomData struct {
			ID           string `json:"id"`
			Title        string `json:"title"`
			BuildingName string `json:"buildingName"`
			BuildingNr   string `json:"buildingNr"`
			CampusName   string `json:"campusName"`
		}
		defer resp.Body.Close()
		str, _ := ioutil.ReadAll(resp.Body)
		_ = json.Unmarshal(str, &roomData)

		checkinPost := struct {
			Location     string `json:"location"`
			LocationName string `json:"locationName"`
			StartTime    int64  `json:"startTime"`
			EndTime      int64  `json:"endTime"`
			SeatNr       string `json:"seatNr"`
			RoomTitle    string `json:"roomTitle"`
			BuildingName string `json:"buildingName"`
			CampusName   string `json:"campusName"`
		}{
			BuildingName: roomData.BuildingName,
			CampusName:   roomData.CampusName,
			EndTime:      to.Unix() * 1000,
			Location:     *roomFlag,
			LocationName: strings.Join([]string{roomData.Title, roomData.BuildingName, roomData.CampusName}, " "),
			RoomTitle:    roomData.Title,
			SeatNr:       "",
			StartTime:    from.Unix() * 1000,
		}

		data, _ := json.Marshal(checkinPost)

		var postResp *http.Response
		if postResp, err = client.Post("https://innsida.ntnu.no/checkin/api/", "application/json", bytes.NewBuffer(data)); err != nil {
			log.Fatal(err)
		}

		defer postResp.Body.Close()
		apiRes, _ := ioutil.ReadAll(postResp.Body)
		fmt.Printf("Checked in to %s from %s to %s: %s\n", checkinPost.LocationName, *startTimestamp, *endTimestamp, apiRes)

		if string(apiRes) != "OK" {
			fmt.Printf(`Status from NTNU checkin is not "OK", so go to https://innsida.ntnu.no/checkin/mycheckins to check: %s\n`, apiRes)
			os.Exit(1)
		}

	case "delete":
		if len(os.Args) < 3 {
			fmt.Printf("Missing id to delete.\n")
			os.Exit(1)
		}
		id := os.Args[2]

		req, err := http.NewRequest("DELETE", "https://innsida.ntnu.no/checkin/api/checkin/"+id, nil)
		var deleteResp *http.Response
		if deleteResp, err = client.Do(req); err != nil {
			log.Fatal(err)
		}

		defer deleteResp.Body.Close()
		apiRes, _ := ioutil.ReadAll(deleteResp.Body)
		fmt.Printf("Deleted checkin %s: %s\n", id, apiRes)

		if string(apiRes) != "Checkin deleted." {
			fmt.Printf(`Status from NTNU checkin is not "Checkin deleted.", so go to https://innsida.ntnu.no/checkin/mycheckins to check: %s\n`, apiRes)
			os.Exit(1)
		}

	case "list":
		fallthrough
	case "get":
		var listResp *http.Response
		if listResp, err = client.Get("https://innsida.ntnu.no/checkin/api/checkinhistory?start=0&end=2000000000000"); err != nil {
			log.Fatal(err)
		}
		var listData []struct {
			ID           string    `json:"_id"`
			Location     string    `json:"location"`
			LocationName string    `json:"locationName"`
			StartTime    time.Time `json:"startTime"`
			EndTime      time.Time `json:"endTime"`
			SeatNr       string    `json:"seatNr"`
			User         string    `json:"user"`
			CreatedAt    time.Time `json:"createdAt"`
			UpdatedAt    time.Time `json:"updatedAt"`
			V            int       `json:"__v"`
		}
		defer listResp.Body.Close()
		str9, _ := ioutil.ReadAll(listResp.Body)
		_ = json.Unmarshal(str9, &listData)
		fmt.Printf("CHECKIN-ID                START              END                ROOM-ID      LOCATION\n")

		now := time.Now()
		for _, checkin := range listData {
			fmt.Printf("%-25s %-18s %-18s %-12s %s\n", checkin.ID, checkin.StartTime.In(now.Location()).Format("Mon Jan _2 15:04"), checkin.EndTime.In(now.Location()).Format("Mon Jan _2 15:04"), checkin.Location, checkin.LocationName)
		}
	case "search":
		var resp9 *http.Response
		if resp9, err = client.Get("https://innsida.ntnu.no/checkin/api/search?query=" + url.QueryEscape(strings.Join(os.Args[2:], " "))); err != nil {
			log.Fatal(err)
		}
		var searchData struct {
			Docs []struct {
				ID           string `json:"id"`
				Title        string `json:"title"`
				BuildingName string `json:"buildingName"`
				CampusName   string `json:"campusName"`
			} `json:"docs"`
		}
		defer resp9.Body.Close()
		str9, _ := ioutil.ReadAll(resp9.Body)
		_ = json.Unmarshal(str9, &searchData)
		fmt.Printf("ROOM-ID    NAME\n")
		for _, room := range searchData.Docs {
			fmt.Printf("%-10s %s\n", strings.Replace(room.ID, "room_", "", 1), strings.Join([]string{room.Title, room.BuildingName, room.CampusName}, " "))
		}
	default:
		fmt.Printf("Unknown command\n")
		help()
		os.Exit(1)
	}
}
