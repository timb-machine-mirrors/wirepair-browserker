package scanner

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/scanner/browser"
	"gitlab.com/browserker/scanner/crawler"
	"gitlab.com/browserker/scanner/report"
)

// Browserk is our engine
type Browserk struct {
	cfg         *browserk.Config
	attackGraph browserk.AttackGrapher
	crawlGraph  browserk.CrawlGrapher
	reporter    browserk.Reporter
	browsers    browser.BrowserPool

	scanContext  context.Context
	stateMonitor *time.Ticker
}

// New engine
func New(cfg *browserk.Config, crawl browserk.CrawlGrapher, attack browserk.AttackGrapher) *Browserk {
	return &Browserk{cfg: cfg, attackGraph: attack, crawlGraph: crawl, reporter: report.New()}
}

// SetReporter overrides the default reporter
func (b *Browserk) SetReporter(reporter browserk.Reporter) *Browserk {
	b.reporter = reporter
	return b
}

// Init the browsers and stores
func (b *Browserk) Init(ctx context.Context) error {
	b.scanContext = ctx

	log.Logger.Info().Msg("initializing attack graph")
	if err := b.attackGraph.Init(); err != nil {
		return err
	}

	log.Logger.Info().Msg("initializing crawl graph")
	if err := b.crawlGraph.Init(); err != nil {
		return err
	}
	b.stateMonitor = time.NewTicker(time.Duration(time.Second * 30))

	log.Logger.Info().Msg("starting leaser")
	leaser := browser.NewLocalLeaser()
	log.Logger.Info().Msg("leaser started")
	pool := browser.NewGCDBrowserPool(b.cfg.NumBrowsers, leaser)
	b.browsers = pool
	log.Logger.Info().Msg("starting browser pool")
	return pool.Init()
}

// Start the browsers
func (b *Browserk) Start() error {
	for {
		select {
		case <-b.scanContext.Done():
			log.Info().Msg("scan finished due to context complete")
			return nil
		case <-b.stateMonitor.C:
			// TODO: check graph for inprocess values that never made it
			log.Info().Msg("state monitor ping")
		default:
			entries := b.crawlGraph.Find(b.scanContext, browserk.NavUnvisited, browserk.NavInProcess, b.cfg.NumBrowsers)

		}
	}
	return nil //b.browsers.Load(context.Background(), b.cfg.URL)
}

func (b *Browserk) processEntries(entries [][]*browserk.Navigation) {
	for i, navs := range entries {
		browser, err := b.browsers.Take(ctx)
		if err != nil {
			return
		}
		crawler := crawler.New(b.cfg)
		if err := crawler.Init(); err != nil {
			return err
		}

		for j, nav := range navs {
			ctx, cancel := context.WithTimeout(b.scanContext, time.Second*45)
			defer cancel()
			results, newEntries, err := crawler.Process(ctx, browser, nav)
		}

	}
}

// Stop the browsers
func (b *Browserk) Stop() error {
	return nil
}
