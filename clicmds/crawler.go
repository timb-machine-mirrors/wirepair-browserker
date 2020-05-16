package clicmds

import (
	"context"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/pelletier/go-toml"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/scanner"
	"gitlab.com/browserker/store"
)

func CrawlerFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "url",
			Usage: "url as a start point",
			Value: "http://localhost/",
		},
		&cli.StringFlag{
			Name:  "config",
			Usage: "config to use",
			Value: "",
		},
		&cli.StringFlag{
			Name:  "datadir",
			Usage: "data directory",
			Value: "browserktmp",
		},
	}
}

func Crawler(ctx *cli.Context) error {
	cfg := &browserk.Config{}
	if ctx.String("config") == "" {
		cfg = &browserk.Config{
			URL:           ctx.String("url"),
			AllowedHosts:  nil,
			ExcludedHosts: nil,
			DataPath:      "",
			AuthScript:    "",
			AuthType:      0,
			Credentials: &browserk.Credentials{
				Username: "",
				Password: "",
				Email:    "",
			},
			NumBrowsers: 3,
		}
	} else {
		data, err := ioutil.ReadFile(ctx.String("config"))
		if err != nil {
			return err
		}
		toml.Unmarshal(data, cfg)
		if cfg.URL == "" && ctx.String("url") != "" {
			cfg.URL = ctx.String("url")
		}
		if cfg.DataPath == "" && ctx.String("datadir") != "" {
			cfg.DataPath = ctx.String("datadir")
		}
	}

	os.RemoveAll(cfg.DataPath)
	crawl := store.NewCrawlGraph(cfg.DataPath + "/crawl")
	attack := store.NewAttackGraph(cfg.DataPath + "/attack")
	browserk := scanner.New(cfg, crawl, attack)
	log.Logger.Info().Msg("Starting browserker")

	scanContext := context.Background()
	if err := browserk.Init(scanContext); err != nil {
		log.Logger.Error().Err(err).Msg("failed to init engine")
		return err
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Info().Msg("Ctrl-C Pressed, shutting down")
		err := browserk.Stop()
		if err != nil {
			log.Error().Err(err).Msg("failed to stop browserk")
		}
		os.Exit(1)
	}()

	err := browserk.Start()
	if err != nil {
		log.Error().Err(err).Msg("browserk failure occurred")
	}

	return browserk.Stop()
}
