package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/gocolly/colly"
)

type Data struct {
	GameID      string
	GameName    string
	TotalPages  int
	WorkshopIDs []int
	StartPage   int
	EndPage     int
	Delay       int
	RDelay      int
}

func parseFlags(data *Data) {

	problem := false

	gameID := flag.String("gameID", "", "ID of the game")
	startPage := flag.Int("startPage", 1, "which workshop page to start at")
	endPage := flag.Int("endPage", 0, "set which workshop page to end at")
	delay := flag.Int("delay", 25, "set the delay between each request (in milliseconds)")
	randomDelay := flag.Int("randomDelay", 0, "add an extra randomized duration to wait added to Delay before creating a new request (in milliseconds)")

	flag.Parse()

	// Check if the gameID flag was provided
	if *gameID == "" {
		fmt.Println("Error: the steam game ID is required")
		problem = true
	}

	data.GameID = *gameID
	data.StartPage = *startPage
	data.EndPage = *endPage
	data.Delay = *delay
	data.RDelay = *randomDelay

	if problem {
		flag.Usage()
		log.Fatalln("There is at least one problem that has to be resolved")
	}

}

func main() {

	var data Data

	parseFlags(&data)

	// Validate the game ID
	if err := checkGame(&data); err != nil {
		log.Fatalf("Problem: '%v'", err)
	}

	log.Printf("found game '%s'", data.GameName)

	getItems(&data)

	if err := saveData(&data); err != nil {
		log.Fatalf("%v\n", err)
	}

	log.Printf("File saved as '%s - %s.txt'", data.GameID, data.GameName)
}

func saveData(data *Data) error {
	// Open the file for writing. Create the file if it doesn't exist. Truncate the file if it does exist.

	fileName := fmt.Sprintf("%s - %s.txt", data.GameID, data.GameName)

	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a new buffered writer
	writer := bufio.NewWriter(file)

	// Write each WorkshopID to the file, each on a new line
	for _, id := range data.WorkshopIDs {
		_, err := writer.WriteString(fmt.Sprintf("%d\n", id))
		if err != nil {
			return err
		}
	}

	// Flush any buffered data to the file
	err = writer.Flush()
	if err != nil {
		return err
	}

	return nil
}

func getItems(data *Data) error {

	// https://steamcommunity.com/workshop/browse/?appid=108600&browsesort=toprated&section=readytouseitems&actualsort=toprated&p=1

	c := colly.NewCollector()

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Delay:       time.Duration(data.Delay) * time.Millisecond,
		RandomDelay: time.Duration(data.RDelay) * time.Millisecond, // Up to this deplay will be added randomly to the set delay
	})

	c.OnHTML("div.workshopBrowseItems > div", func(e *colly.HTMLElement) {
		e.ForEach("a.ugc", func(i int, h *colly.HTMLElement) {

			parsedURL, err := url.Parse(h.Attr("href"))
			if err != nil {
				log.Printf("Error parsing URL: %v\n", err)
				return
			}

			queryParams := parsedURL.Query()

			id := queryParams.Get("id")

			IDasInt, err := strconv.Atoi(id)
			if err != nil {
				log.Printf("Error parsing integer: %v\n", err)
				return
			}

			data.WorkshopIDs = append(data.WorkshopIDs, IDasInt)
		})
	})

	iterateTo := data.TotalPages
	if data.EndPage > 0 {
		iterateTo = data.EndPage
	}

	for i := data.StartPage; i <= iterateTo; i++ {

		url := fmt.Sprintf("https://steamcommunity.com/workshop/browse/?appid=%s&browsesort=toprated&section=readytouseitems&actualsort=toprated&p=%d", data.GameID, i)
		fmt.Printf("Visiting page %d / %d\n", i, iterateTo)
		c.Visit(url)
	}

	if err := c.Visit(fmt.Sprintf("https://steamcommunity.com/app/%s/workshop/", data.GameID)); err != nil {
		return err
	}
	return nil
}

func checkGame(data *Data) error {

	ok := true

	c := colly.NewCollector()
	c.OnResponse(func(r *colly.Response) {
		//pattern := `^https:\/\/steamcommunity\.com\/app\/\d+\/workshop`
		pattern := `^https:\/\/steamcommunity\.com\/workshop\/browse\/\?appid=\d+`

		re, _ := regexp.Compile(pattern)

		if !re.MatchString(r.Request.URL.String()) {
			ok = false
		}
	})

	c.OnHTML("div.apphub_AppName.ellipsis", func(e *colly.HTMLElement) {
		data.GameName = e.Text
	})

	c.OnHTML("div.workshopBrowsePagingControls > a:nth-last-child(2)", func(e *colly.HTMLElement) {

		IDasInt, err := strconv.Atoi(e.Text)
		if err != nil {
			log.Printf("Error parsing page integer: %v\n", err)
			ok = false
			return
		}

		data.TotalPages = IDasInt
	})

	//url := fmt.Sprintf("https://steamcommunity.com/app/%s/workshop/", data.GameID)
	url := fmt.Sprintf("https://steamcommunity.com/workshop/browse/?appid=%s&browsesort=toprated&section=readytouseitems", data.GameID)

	if err := c.Visit(url); err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("could not find game, with ID '%s', on Steam", data.GameID)
	}

	return nil
}
