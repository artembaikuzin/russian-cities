package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gocolly/colly/v2"
)

type City struct {
	Name   string
	Region string
	Lat    float64
	Lon    float64
}

func main() {
	start := time.Now()

	var fixRegions = flag.Bool("fix-regions", false, "Tatarstan -> Republic of Tatarstan, and no federal cities")
	flag.Parse()

	fmt.Fprintln(os.Stderr, "🌏 Scraping...")

	defer func() {
		fmt.Fprintln(os.Stderr, "Took", time.Since(start))
	}()

	baseUrl := "https://ru.wikipedia.org"
	url := "https://ru.wikipedia.org/wiki/%D0%A1%D0%BF%D0%B8%D1%81%D0%BE%D0%BA_%D0%B3%D0%BE%D1%80%D0%BE%D0%B4%D0%BE%D0%B2_%D0%A0%D0%BE%D1%81%D1%81%D0%B8%D0%B8"
	cities := 0

	c := colly.NewCollector()

	c.OnHTML("table tbody tr", func(e *colly.HTMLElement) {
		if len(e.DOM.Children().Nodes) != 9 {
			return
		}

		city := City{}

		e.ForEach("td", func(i int, h *colly.HTMLElement) {
			switch i {
			case 2:
				h.ForEach("a:first-child", func(i int, h *colly.HTMLElement) {
					if i > 0 {
						return
					}

					city.Name = h.Text
					cityPage := h.Attr("href")

					d := c.Clone()

					d.OnHTML("span.coordinates a.mw-kartographer-maplink", func(h *colly.HTMLElement) {
						if city.Lat > 0.0 {
							return
						}

						lat, err := strconv.ParseFloat(h.Attr("data-lat"), 64)

						if err != nil {
							fmt.Fprintf(os.Stderr, "Can't parse lat for %q, %v\n", city.Name, err)
							return
						}

						lon, err := strconv.ParseFloat(h.Attr("data-lon"), 64)

						if err != nil {
							fmt.Fprintf(os.Stderr, "Can't parse lon for %q, %v\n", city.Name, err)
							return
						}

						city.Lat = lat
						city.Lon = lon
					})

					d.Visit(baseUrl + cityPage)
				})
			case 3:
				if *fixRegions {
					city.Region = fixRegion(h.Text)
				} else {
					city.Region = h.Text
				}

				cities += 1

				fmt.Printf("%s,%s,%f,%f\n", city.Name, city.Region, city.Lat, city.Lon)
			}
		})
	})

	c.Visit(url)

	fmt.Fprintln(os.Stderr, "Total cities", cities)
}

func fixRegion(region string) string {
	translate := map[string]string{
		"Москва":          "Московская область",
		"Санкт-Петербург": "Ленинградская область",
		"Севастополь":     "Республика Крым",

		"Адыгея":             "Республика Адыгея",
		"Алтай":              "Республика Алтай",
		"Башкортостан":       "Республика Башкортостан",
		"Бурятия":            "Республика Бурятия",
		"Дагестан":           "Республика Дагестан",
		"Ингушетия":          "Республика Ингушетия",
		"Кабардино-Балкария": "Кабардино-Балкарская Республика",
		"Калмыкия":           "Республика Калмыкия",
		"Карачаево-Черкесия": "Карачаево-Черкесская Республика",
		"Карелия":            "Республика Карелия",
		"Коми":               "Республика Коми",
		"Крым":               "Республика Крым",
		"Марий Эл":           "Республика Марий Эл",
		"Мордовия":           "Республика Мордовия",
		"Северная Осетия":    "Республика Северная Осетия–Алания",
		"Татарстан":          "Республика Татарстан",
		"Тыва":               "Республика Тыва",
		"Удмуртия":           "Удмуртская Республика",
		"Хакасия":            "Республика Хакасия",
		"Чечня":              "Чеченская Республика",
		"Чувашия":            "Чувашская Республика",
		"Якутия":             "Республика Саха",
	}

	t, ok := translate[region]

	if !ok {
		return region
	}

	return t
}
