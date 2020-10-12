package main

import (
	"github.com/labstack/echo"
	"os"
	"strings"
	"webscrapper/internal"
)

const fileName string = "jobs.csv"

func handleHome(c echo.Context) error{
	return c.File("../web/home.html")
}
func handleScrape(c echo.Context) error{
	defer os.Remove(fileName)
	term := strings.ToLower(scrapper.CleanString(c.FormValue("term")))
	scrapper.Scrape(term)
	return c.Attachment(fileName, fileName)
}
func main(){
	e:= echo.New()
	e.GET("/", handleHome)
	e.POST("/scrape", handleScrape)
	e.Logger.Fatal(e.Start(":1323"))
}