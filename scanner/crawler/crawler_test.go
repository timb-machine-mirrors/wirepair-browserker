package crawler_test

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/mock"
	"gitlab.com/browserker/scanner"
	"gitlab.com/browserker/scanner/browser"
	"gitlab.com/browserker/scanner/crawler"
)

type crawlerTests struct {
	formHandler func(c *gin.Context)
	url         string
}

var leaser = browser.NewLocalLeaser()

func testServer(path string, fn gin.HandlerFunc) (string, *http.Server) {
	router := gin.Default()
	router.Static("/forms", "testdata/forms")
	if fn != nil {
		router.Any(path, fn)
	}
	testListener, _ := net.Listen("tcp", ":0")
	_, testServerPort, _ := net.SplitHostPort(testListener.Addr().String())
	srv := &http.Server{
		Addr:    testListener.Addr().String(),
		Handler: router,
	}
	//testServerAddr := fmt.Sprintf("http://localhost:%s/", testServerPort)
	go func() {
		if err := srv.Serve(testListener); err != http.ErrServerClosed {
			log.Fatalf("Serve(): %s", err)
		}
	}()

	return testServerPort, srv
}

func TestCrawler(t *testing.T) {
	pool := browser.NewGCDBrowserPool(1, leaser)
	if err := pool.Init(); err != nil {
		t.Fatalf("failed to init pool")
	}
	defer leaser.Cleanup()
	ctx := context.Background()
	bCtx := mock.Context(ctx)
	bCtx.Log = &zerolog.Logger{}
	bCtx.FormHandler = crawler.NewCrawlerFormHandler(&browserk.DefaultFormValues)

	called := false

	simpleCallFunc := func(c *gin.Context) {
		called = true
		resp := "<html><body>You made it!</body></html>"
		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Write([]byte(resp))
	}

	toTest := [...]crawlerTests{
		{
			func(c *gin.Context) {
				fname, _ := c.GetQuery("fname")
				lname, _ := c.GetQuery("lname")

				if fname == "Test" && lname == "User" {
					called = true
				}

				resp := "<html><body>You made it!</body></html>"
				c.Writer.WriteHeader(http.StatusOK)
				c.Writer.Write([]byte(resp))
			},
			"http://localhost:%s/forms/",
		},
		{
			func(c *gin.Context) {
				fname, _ := c.GetQuery("fname")
				lname, _ := c.GetQuery("lname")
				car, _ := c.GetQuery("cars")

				if fname == "Test" && lname == "User" && car == "volvo" {
					called = true
				}

				resp := "<html><body>You made it!</body></html>"
				c.Writer.WriteHeader(http.StatusOK)
				c.Writer.Write([]byte(resp))
			},
			"http://localhost:%s/forms/select.html",
		},
		{
			func(c *gin.Context) {
				fname, _ := c.GetQuery("fname")
				lname, _ := c.GetQuery("lname")
				rad, _ := c.GetQuery("rad")

				if fname == "Test" && lname == "User" && rad == "rad1" {
					called = true
				}

				resp := "<html><body>You made it!</body></html>"
				c.Writer.WriteHeader(http.StatusOK)
				c.Writer.Write([]byte(resp))
			},
			"http://localhost:%s/forms/radio.html",
		},
		{
			simpleCallFunc,
			"http://localhost:%s/forms/onmouseclick.html",
		},
		{
			simpleCallFunc,
			"http://localhost:%s/forms/onmousedblclick.html",
		},
		{
			simpleCallFunc,
			"http://localhost:%s/forms/onmousedown.html",
		},
		{
			simpleCallFunc,
			"http://localhost:%s/forms/onmouseenter.html",
		},
		{
			simpleCallFunc,
			"http://localhost:%s/forms/onmouseleave.html",
		},
		{
			simpleCallFunc,
			"http://localhost:%s/forms/onmouseout.html",
		},
		{
			simpleCallFunc,
			"http://localhost:%s/forms/onmouseup.html",
		},
		{
			simpleCallFunc,
			"http://localhost:%s/forms/keydown.html",
		},
		{
			simpleCallFunc,
			"http://localhost:%s/forms/keypress.html",
		},
		{
			simpleCallFunc,
			"http://localhost:%s/forms/keyup.html",
		},
	}

	for _, crawlTest := range toTest {
		b, port, err := pool.Take(bCtx)
		if err != nil {
			t.Fatalf("error taking browser: %s\n", err)
		}
		p, srv := testServer("/result/formResult", crawlTest.formHandler)
		defer srv.Shutdown(ctx)
		target := fmt.Sprintf(crawlTest.url, p)
		targetURL, _ := url.Parse(target)
		bCtx.Scope = scanner.NewScopeService(targetURL)
		crawl := crawler.New(&browserk.Config{})
		t.Logf("going to %s\n", target)
		act := browserk.NewLoadURLAction(target)
		nav := browserk.NewNavigation(browserk.TrigCrawler, act)
		_, newNavs, err := crawl.Process(bCtx, b, nav, true)
		if err != nil {
			t.Fatalf("error getting url %s\n", err)
		}

		if len(newNavs) != 1 {
			t.Fatal("did not find form nav action")
		}

		_, _, err = crawl.Process(bCtx, b, newNavs[0], true)
		if err != nil {
			t.Fatalf("failed to submit form %s\n", err)
		}

		if !called {
			t.Fatalf("form was not submitted: %s\n", target)
		}
		called = false
		pool.Return(ctx, port)
		srv.Shutdown(ctx)
	}

}
