package scrapper

import (
	"encoding/csv"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type extractedJob struct{
	link string
	title string
	company string
	location string
	salary string
	summary string
}

func Scrape(searchWord string){
	var baseURL = "https://kr.indeed.com/"
	URL := baseURL+"jobs?q="+searchWord+"&limit=50"
	var jobs []extractedJob
	c:= make(chan []extractedJob)

	totalPages := getTotalPages(URL)

	for i:=0;i<totalPages;i++{
		go getPage(URL, i, c)
	}
	for i:=0;i<totalPages;i++{
		jobs = append(jobs, <-c...)
	}
	writeFile(jobs)
	fmt.Println("Done, extracted", len(jobs))
}

// getTotalPages from recruitment site
func getTotalPages (URL string) int{
	totalPages := 0
	res, err := http.Get(URL)
	checkErr(err)
	checkStatusCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	doc.Find(".pagination").Each(func(i int, s *goquery.Selection){
		totalPages = s.Find("a").Length()
	})

	if totalPages == 0 {
		return 1
	}
	return totalPages
}

// getPage from recruitment site
func getPage(baseURL string, page int, mainC chan<- []extractedJob) {
	var jobs []extractedJob
	c:= make(chan extractedJob)

	URL := baseURL +"&start="+strconv.Itoa(page*50)
	fmt.Println("Requesting", URL)
	res, err:= http.Get(URL)
	checkErr(err)
	checkStatusCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	searchCards := doc.Find(".jobsearch-SerpJobCard")
	searchCards.Each(func(i int, card *goquery.Selection){
		go extractJob(card, c)
	})

	for i:=0;i<searchCards.Length();i++{
		jobs = append(jobs, <-c)
	}
	mainC <- jobs
}

// extractJob detail from recruitment site
func extractJob(card *goquery.Selection, c chan<- extractedJob) {
	id,_ := card.Attr("data-jk")
	title := CleanString(card.Find(".title>a").Text())
	company := CleanString(card.Find(".company").Text())
	location := CleanString(card.Find(".location").Text())
	salary := CleanString(card.Find("salary").Text())
	summary := CleanString(card.Find(".summary").Text())
	c <- extractedJob{"https://kr.indeed.com/viewjob?jk="+id,title,company,location,salary,summary}
}
func writeFile(jobs []extractedJob){
	c := make(chan []string)

	file, err := os.Create("jobs.csv")
	checkErr(err)
	w := csv.NewWriter(file)

	defer w.Flush()

	headers := []string{"Link", "Title", "Company", "Location", "Salary", "Summary"}
	wErr:= w.Write(headers)
	checkErr(wErr)

	for _, job := range jobs{
		go writeJob(job, c)
	}
	for i:=0;i<len(jobs);i++{
		wErr := w.Write(<-c)
		checkErr(wErr)
	}
}
func writeJob(job extractedJob, c chan<- []string){
	c<- []string{job.link, job.title, job.company, job.location, job.salary, job.summary}
}
func checkErr (err error) {
	if err != nil{
		log.Fatalln(err)
	}
}
func checkStatusCode (res *http.Response){
	if res.StatusCode != 200 {
		log.Fatalln("Request failed with Status :", res.StatusCode)
	}
}
func CleanString (str string) string{
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}