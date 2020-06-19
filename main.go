package main

import (
	"bytes"
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/btnmasher/lumberjack"
	"github.com/kr/pretty"
)

var (
	user   string
	repo   string
	text   string
	dryrun bool
	usessh bool
	logger *lumberjack.Logger
)

func init() {
	usr := flag.String("user", "", "Github username.")
	rep := flag.String("repo", "", "Name of the repo to generate.")
	txt := flag.String("text", "", "Text to set for the histogram.")
	dbg := flag.Bool("debug", false, "Turns on debug loggerging.")
	dry := flag.Bool("dryrun", false, "Disables output of generated bash script, only histogram preview.")
	ssh := flag.Bool("ssh", false, "Enables output of script using git with SSH instead of the default HTTPS.")
	flag.Parse()

	user = *usr
	repo = *rep
	text = *txt
	dryrun = *dry
	usessh = *ssh

	logger = lumberjack.NewLoggerWithDefaults()

	if *dbg {
		logger.AddLevel(lumberjack.DEBUG)
	}
}

func main() {
	resp, err := http.Get(fmt.Sprintf("https://github.com/users/%s/contributions", user))
	if err != nil {
		logger.Fatalf("Could not retrieve github data: %s", err)
	}

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		logger.Fatal("Could not parse returned XML: %s", err)
	}

	startdate := getFirstFullDate(doc)
	logger.Debugf("Got first available full-column date: %s", startdate)

	multiplier := getHighestCommitCount(doc) / 3 * 2
	logger.Debugf("Multiplier set to: %v", multiplier)

	//Get the letter arrays from the string
	symbols := StringToPixelSymbols(text)
	logger.Debugf("Converted string %s to symbols: \n%s", text, pretty.Sprintf("%s", symbols))

	//Build and print the script to stdout
	fmt.Println(buildScript(user, repo, text, multiplier, startdate, symbols))
}

func buildScript(user, repo, text string, multi int, startdate time.Time, symbols []Symbol) string {
	var buffer bytes.Buffer

	header := "#!/bin/bash\ngit init %s\ncd %s\ntouch README.md\ngit add README.md\n"
	commit := "GIT_AUTHOR_DATE=%s GIT_COMMITTER_DATE=%s git commit --allow-empty -m \"Rewriting History!\" > /dev/null\n"
	footerhttps := "git remote add origin https://github.com/%s/%s.git\ngit pull\ngit push -u origin master\n"
	footerssh := "git remote add origin git@github.com:%s/%s.git\ngit pull\ngit push -u origin master\n"
	timeformat := "2006-01-02T15:04:05"

	//Initialize our histogram buffer
	var pixels [][]byte
	pixels = make([][]byte, 7, 7)
	for i := range pixels {
		pixels[i] = make([]byte, 52, 52)
	}

	logger.Debugf("Built empty histogram buffer:\n%s", printHistogram(pixels))

	//Give us 1 column offset.
	curcol := 1
	startdate = startdate.Add((time.Hour * 24) * 7)

	//Start rendering the histogram
	for _, l := range symbols {
		logger.Debugf("Handling symbol: %s", l.Rune)
		if curcol+l.Width() < 52 {
			for r := 0; r < 7; r++ {
				for c := 0; c < 52-curcol; c++ {
					if c < l.Width() {
						pixels[r][c+curcol] = l.Pixels[r][c]
					} else {
						break
					}
				}
			}
		} else {
			logger.Debugf("Symbol overlapped end of available space at column: %v", curcol)
			break //Ran out of room
		}
		curcol += l.Width()
	}

	if !dryrun {
		logger.Debug("Generating script output.")

		//Write out our header line.
		buffer.WriteString(fmt.Sprintf(header, repo, repo))
	}

	//Print that beautiful thing
	buffer.WriteString(printHistogram(pixels))

	if !dryrun {
		//Generate the commits
		for c := 0; c < 52; c++ {
			for r := 0; r < 7; r++ {
				if pixels[r][c] != 0 {
					date := startdate.Format(timeformat)
					str := fmt.Sprintf(commit, date, date)
					for i := 0; i < int(pixels[r][c])*multi; i++ {
						buffer.WriteString(str)
					}
				}
				startdate = startdate.Add(time.Hour * 24)
			}
		}

		//Finish the script
		if usessh {
			buffer.WriteString(fmt.Sprintf(footerssh, user, repo))
		} else {
			buffer.WriteString(fmt.Sprintf(footerhttps, user, repo))
		}

		logger.Debug("Script generated!")
	} else {
		logger.Debug("Skipping script output, dryrun=true")
	}

	return buffer.String()
}

func printHistogram(pixels [][]byte) string {
	var pbuf bytes.Buffer

	pbuf.WriteString("\n# ----- Histogram Preview -----\n\n")

	for r := 0; r < 7; r++ {
		pbuf.WriteString("#    ")
		for c := 0; c < 52; c++ {
			if r < len(pixels[r]) {
				b := pixels[r][c]
				if r, exists := PixelRune[b]; exists {
					pbuf.WriteRune(r)
				} else {
					pbuf.WriteRune('!') //Unknown rune value
					logger.Errorf("Encountered invalid pixel value byte: %v", b)
				}
			} else {
				pbuf.WriteRune('!') //Unallocated pixel
				logger.Errorf("Encountered unallocated index in pixel value byte array. Row: %v - Col: %v", r, c)
			}
		}
		pbuf.WriteRune('\n')
	}

	pbuf.WriteRune('\n')

	return pbuf.String()
}

func getFirstFullDate(doc *goquery.Document) time.Time {
	var firstdate time.Time
	doc.Find("g").Each(func(i int, s *goquery.Selection) {
		t, ok := s.Attr("transform")
		if ok {
			if t == "translate(16, 0)" {
				d, ok := s.Find(".day").First().Attr("data-date")
				if ok {
					var err error
					firstdate, err = time.Parse("2006-01-02", d)
					if err != nil {
						logger.Fatalf("Error parsing date: %s", err)
					}
					return
				}
			}
		}
	})
	return firstdate
}

func getHighestCommitCount(doc *goquery.Document) int {
	var highest int

	doc.Find(".day").Each(func(i int, s *goquery.Selection) {
		count, ok := s.Attr("data-count")
		if ok {
			if n, err := strconv.Atoi(count); err != nil {
				logger.Fatalf("Error parsing count from data: %s", err)
			} else {
				if n > highest {
					highest = n
				}
			}
		}
	})
	return highest
}
