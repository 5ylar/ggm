package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/playwright-community/playwright-go"
)

func main() {
	err := playwright.Install()

	if err != nil {
		panic(err)
	}

	pw, err := playwright.Run()

	if err != nil {
		panic(err)
	}

	defer pw.Stop()

	browser, err := pw.Chromium.Launch(
		playwright.BrowserTypeLaunchOptions{
			Headless: playwright.Bool(false),
			// ChromiumSandbox: playwright.Bool(true),
			// IgnoreAllDefaultArgs: playwright.Bool(true), //
			// Args: []string{
			// "--disable-gpu",
			// 	"--disable-gpu-driver-bug-workarounds",
			// 	"--disable-gpu-sandbox",
			// 	"--disable-gpu-vsync",
			// 	"--disable-gpu-early-init",
			// 	"--disable-gpu-program-cache",
			// 	"--disable-gpu-compositing",
			// 	"--disable-software-rasterizer",
			// 	"--allow-pre-commit-input",
			// 	"--disable-background-networking",
			// 	"--disable-background-timer-throttling",
			// 	"--disable-backgrounding-occluded-windows",
			// 	"--disable-breakpad",
			// 	"--disable-client-side-phishing-detection",
			// 	"--disable-component-update",
			// 	"--disable-dev-shm-usage",
			// 	"--disable-field-trial-config",
			// 	"--disable-hang-monitor",
			// 	"--disable-infobars",
			// 	"--disable-ipc-flooding-protection",
			// 	"--disable-popup-blocking",
			// 	"--disable-prompt-on-repost",
			// 	"--disable-renderer-backgrounding",
			// 	"--disable-search-engine-choice-screen",
			// 	"--disable-sync",
			// 	// "--enable-automation",
			// 	// "--export-tagged-pdf",
			// 	"--force-color-profile=srgb",
			// 	// "--metrics-recording-only",
			// 	"--no-first-run",
			// 	"--password-store=basic",
			// 	"--use-mock-keychain",
			// 	"--disable-features=Translate,AcceptCHFrame,MediaRouter,OptimizationHints,ProcessPerSiteUpToMainFrameThreshold",
			// 	// "--enable-features=NetworkServiceInProcess2",
			// 	"--disable-blink-features=AutomationControlled",
			// 	// "about:blank",
			// 	// "--remote-debugging-port=9222",
			// 	// "--remote-allow-origins='*'",
			// 	// "--remote-debugging-address=0.0.0.0",
			// 	// "--guest",
			// },
		},
	)
	// browser, err := pw.Chromium.ConnectOverCDP("http://0.0.0.0:9222")

	if err != nil {
		panic(err)
	}

	defer browser.Close()

	blockedURLs := []string{
		"https://fonts.gstatic.com/",
		"https://maps.gstatic.com/",
		"https://www.google.co.th/maps/preview/log",
		// "https://www.google.co.th/maps/vt/stream/pb=",
		"ogads-pa.clients6.google.com/$rpc/google.internal.onegoogle.asyncdata.v1.AsyncDataService/GetAsyncData",
		"data:image/png;base64,",
		"https://lh5.googleusercontent.com/p/",
		"https://www.google.co.th/maps/rpc/locationsharing/",
		// "https://www.google.co.th/maps/preview/pegman",
		"https://play.google.com/log",
		// "https://www.google.co.th/gen_",
		"https://www.google.com/images/branding/product/ico/",
		"https://www.google.co.th/maps/vt/icon/",
		"https://www.google.co.th/maps/preview/lp",
	}

	// lat := 13.9385606
	// lng := 100.5608079
	// query := "Restaurants"
	// lang := "th"
	// zoom := 18

	// pos := [][]float64{
	// 	{13.8321065, 100.579688},
	// 	{13.8321065, 100.579688},
	// 	{13.8321065, 100.579688},
	// 	{13.8321065, 100.579688},
	// 	{13.8321065, 100.579688},
	// 	{13.8321065, 100.579688},
	// 	{13.8321065, 100.579688},
	// 	{13.8321065, 100.579688},
	// 	{13.8321065, 100.579688},
	// 	{13.8321065, 100.579688},
	// }

	cf, err := os.Open("points.csv")

	if err != nil {
		panic(err)
	}

	cr := csv.NewReader(cf)

	records, err := cr.ReadAll()

	if err != nil {
		panic(err)
	}

	var pos [][]float64

	for i, r := range records {
		// header
		if i == 0 {
			continue
		}

		// 20240705 18:33
		s := 1
		e := 1000

		isTarget := i >= s && i <= e

		if !isTarget {
			continue
		}

		lat, err := strconv.ParseFloat(r[0], 64)

		if err != nil {
			panic(err)
		}

		lon, err := strconv.ParseFloat(r[1], 64)

		if err != nil {
			panic(err)
		}

		pos = append(pos, []float64{lat, lon})
	}

	concurrent := 10

	sem := make(chan struct{}, concurrent)

	linksch := make(chan []string, concurrent)

	var progress atomic.Uint64

	go func() {
		for _, p := range pos {
			sem <- struct{}{}

			go func(p []float64) {

				now := time.Now()

				lat := p[0]
				lng := p[1]
				query := "Restaurants"
				lang := "th"
				zoom := 21

				defer func() {
					<-sem

					if r := recover(); r != nil {
						log.Println("ERROR", lat, lng, query, r)
						linksch <- []string{}
					}
				}()

				browserCtx, err := browser.NewContext()

				if err != nil {
					panic(err)
				}

				defer browserCtx.Close()

				browserCtx.SetDefaultTimeout(1000 * 60 * 10)

				browserCtx.Route("**", func(r playwright.Route) {

					for _, blockedURL := range blockedURLs {
						if strings.HasPrefix(r.Request().URL(), blockedURL) {
							r.Abort()
							return
						}
					}

					r.Continue()
				})

				links := searchByPosition(browserCtx, lat, lng, query, lang, zoom)

				go func() {
					progress.Add(1)

					log.Printf("[%d/%d] %f,%f links len: %d | %f.2s\n", progress.Load(), len(pos), lat, lng, len(links), time.Since(now).Seconds())
				}()

				linksch <- links
			}(p)
		}
	}()

	c := 0

	f, err := os.OpenFile("links.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		panic(err)
	}

	defer f.Close()

	for lc := range linksch {
		if c >= len(pos)-1 {
			break
		}

		for _, l := range lc {
			_, err := f.WriteString("\n" + l)

			if err != nil {
				log.Println("ERROR", err)
			}
		}

		c++
	}
}

// func drag(page playwright.Page, x, y, tx, ty float64) {
// 	// err := page.Mouse().Click(x, y)
//
// 	err := page.Mouse().Move(x, y, playwright.MouseMoveOptions{Steps: func() *int { a := 5; return &a }()})
//
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	err = page.Mouse().Down()
//
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	err = page.Mouse().Move(tx, ty, playwright.MouseMoveOptions{Steps: func() *int { a := 5; return &a }()})
//
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	err = page.Mouse().Up()
//
// 	if err != nil {
// 		panic(err)
// 	}
// }

func searchByPosition(
	browserCtx playwright.BrowserContext,
	lat, lng float64,
	query string,
	lang string,
	zoom int,
) []string {
	page, err := browserCtx.NewPage()

	if err != nil {
		panic(err)
	}

	defer page.Close()

	page.SetDefaultTimeout(1000 * 60 * 10)

	_, err = page.Goto(
		fmt.Sprintf(
			"https://www.google.co.th/maps/search/%s/@%f,%f,%dz?hl=%s&entry=ttu",
			query,
			lat,
			lng,
			zoom,
			lang,
		),
	)

	if err != nil {
		panic(err)
	}

	err = page.WaitForLoadState(
		playwright.PageWaitForLoadStateOptions{
			State: playwright.LoadStateNetworkidle,
		},
	)

	if err != nil {
		panic(err)
	}

	var links []string

	bodyloc := page.Locator("body")

	r, err := bodyloc.Evaluate(`
		const body = document.querySelector('body');
		const cy = body.offsetHeight / 2;
		const cx = body.offsetWidth / 2;
		const r = { cx, cy }
		r
	`, nil)

	cx := float64((r.(map[string]interface{}))["cx"].(int))
	cy := float64((r.(map[string]interface{}))["cy"].(int))
	//
	//
	// time.Sleep(time.Second * 1)
	//
	// drag(page, cx, cy, cx, cy+cy/2)
	//
	// // err = page.Locator("button[jsaction=\"pane.queryOnPan.toggle; focus:pane.queryOnPan.toggle; blur:pane.queryOnPan.toggle; keydown:pane.queryOnPan.toggle\"]").Click()
	// //
	// // if err != nil {
	// // 	panic(err)
	// // }

	page.Mouse().Move(cx, cy)

	page.Mouse().Wheel(0, 50)

	searchselector := "button[jsaction=\"search.refresh\"]"
	//
	// time.Sleep(time.Second * 2)
	//
	// page.WaitForSelector(searchselector)
	//
	// drag(page, cx, cy, cx, cy-cy/2)
	//
	// time.Sleep(time.Second * 3)
	//
	// for {
	// 	count, _ := page.Locator(searchselector).Count()
	//
	// 	if count == 0 {
	// 		break
	// 	}
	//
	err = page.Locator(searchselector).Click()

	if err != nil {
		panic(err)
	}

	//
	err = page.WaitForLoadState(
		playwright.PageWaitForLoadStateOptions{
			State: playwright.LoadStateNetworkidle,
		},
	)

	if err != nil {
		panic(err)
	}
	//
	// 	// err = page.WaitForLoadState(
	// 	// 	playwright.PageWaitForLoadStateOptions{
	// 	// 		State: playwright.LoadStateDomcontentloaded,
	// 	// 	},
	// 	// )
	// 	//
	// 	// if err != nil {
	// 	// 	panic(err)
	// 	// }
	// }

	time.Sleep(time.Second * 3)

	links = append(links, collectLinks(page)...)

	return links
}

func collectLinks(page playwright.Page) []string {
	linksmap := make(map[string]struct{})

	lastaloclen := 0

	feedloc := page.Locator("div[role=feed]")

	for {
		alocs, err := feedloc.Locator("div > div > a").All()

		if err != nil {
			panic(err)
		}

		for _, aloc := range alocs {
			link, err := aloc.GetAttribute("href")

			if err != nil {
				panic(err)
			}

			if len(link) == 0 {
				continue
			}

			if !strings.HasPrefix(link, "https://www.google.co.th/maps/place") {
				continue
			}

			linksmap[link] = struct{}{}
		}

		_, err = feedloc.Evaluate("document.querySelector(\"div[role=feed]\").scrollTo(0, document.querySelector(\"div[role=feed]\").scrollHeight)", nil)

		if err != nil {
			panic(err)
		}

		err = page.WaitForLoadState(
			playwright.PageWaitForLoadStateOptions{
				State: playwright.LoadStateNetworkidle,
			},
		)

		if err != nil {
			panic(err)
		}

		err = page.WaitForLoadState(
			playwright.PageWaitForLoadStateOptions{
				State: playwright.LoadStateDomcontentloaded,
			},
		)

		if err != nil {
			panic(err)
		}

		page.WaitForTimeout(2 * 1000)

		if len(alocs) <= lastaloclen {
			break
		}

		lastaloclen = len(alocs)
	}

	var links []string

	for lm := range linksmap {
		links = append(links, lm)
	}

	return links
}
