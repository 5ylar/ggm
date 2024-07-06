package main

import (
	"fmt"
	"log"
	"strings"
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
		log.Fatalf("could not start playwright: %v", err)
	}

	// lat := 13.9385606
	// lng := 100.5608079
	// query := "Restaurants"
	// lang := "th"
	// zoom := 18

	pos := [][]float64{
		{13.8321065, 100.579688},
		{13.8321065, 100.579688},
		{13.8321065, 100.579688},
		{13.8321065, 100.579688},
		{13.8321065, 100.579688},
		{13.8321065, 100.579688},
		{13.8321065, 100.579688},
		{13.8321065, 100.579688},
		{13.8321065, 100.579688},
		{13.8321065, 100.579688},
	}

	sem := make(chan struct{}, 3)

	linksch := make(chan []string, 3)

	go func() {
		for _, p := range pos {
			sem <- struct{}{}

			go func(p []float64) {
				defer func() {
					<-sem

					if r := recover(); r != nil {
						log.Println("ERROR", r)
						linksch <- []string{}
					}
				}()

				lat := p[0]
				lng := p[1]
				query := "Restaurants"
				lang := "th"
				zoom := 21

				links := searchByPosition(pw, lat, lng, query, lang, zoom)

				log.Println("links 1", len(links))

				linksch <- links
			}(p)
		}
	}()

	c := 0

	for lc := range linksch {

		_ = lc

		c++

		if c >= len(pos) {
			break
		}
	}

	err = pw.Stop()

	if err != nil {
		panic(err)
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
	pw *playwright.Playwright,
	lat, lng float64,
	query string,
	lang string,
	zoom int,
) []string {
	browser, err := pw.Chromium.Launch(
		playwright.BrowserTypeLaunchOptions{
			Headless: playwright.Bool(true),
			// ChromiumSandbox: playwright.Bool(true),
			// IgnoreAllDefaultArgs: playwright.Bool(true), //
			Args: []string{
				"--disable-gpu",
				"--allow-pre-commit-input",
				"--disable-background-networking",
				"--disable-background-timer-throttling",
				"--disable-backgrounding-occluded-windows",
				"--disable-breakpad",
				"--disable-client-side-phishing-detection",
				"--disable-component-update",
				"--disable-dev-shm-usage",
				"--disable-field-trial-config",
				"--disable-hang-monitor",
				"--disable-infobars",
				"--disable-ipc-flooding-protection",
				"--disable-popup-blocking",
				"--disable-prompt-on-repost",
				"--disable-renderer-backgrounding",
				"--disable-search-engine-choice-screen",
				"--disable-sync",
				// "--enable-automation",
				// "--export-tagged-pdf",
				"--force-color-profile=srgb",
				// "--metrics-recording-only",
				"--no-first-run",
				"--password-store=basic",
				"--use-mock-keychain",
				"--disable-features=Translate,AcceptCHFrame,MediaRouter,OptimizationHints,ProcessPerSiteUpToMainFrameThreshold",
				// "--enable-features=NetworkServiceInProcess2",
				"--disable-blink-features=AutomationControlled",
				// "about:blank",
				// "--remote-debugging-port=9222",
				// "--remote-allow-origins='*'",
				// "--remote-debugging-address=0.0.0.0",
				"--guest",
			},
		},
	)
	// browser, err := pw.Chromium.ConnectOverCDP("http://0.0.0.0:9222")

	if err != nil {
		panic(err)
	}

	page, err := browser.NewPage()

	if err != nil {
		panic(err)
	}

	page.SetDefaultTimeout(0)

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

	err = browser.Close()

	if err != nil {
		panic(err)
	}

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
