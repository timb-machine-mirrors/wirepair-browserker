package browser_test

import (
	"fmt"
	"testing"

	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/mock"
	"gitlab.com/browserker/scanner/browser"
	"golang.org/x/net/context"
)

func TestActionClick(t *testing.T) {
	pool := browser.NewGCDBrowserPool(1, leaser)
	if err := pool.Init(); err != nil {
		t.Fatalf("failed to init pool")
	}
	defer leaser.Cleanup()

	ctx := context.Background()
	bCtx := mock.Context(ctx)
	p, srv := testServer()
	defer srv.Shutdown(ctx)

	url := fmt.Sprintf("http://localhost:%s/events.html", p)

	b, _, err := pool.Take(bCtx)
	if err != nil {
		t.Fatalf("error taking browser: %s\n", err)
	}

	err = b.Navigate(ctx, url)
	if err != nil {
		t.Fatalf("error getting url %s\n", err)
	}
	eles, err := b.FindElements("button")
	if err != nil {
		t.Fatalf("error getting elements: %s\n", err)
	}

	for _, ele := range eles {
		act := &browserk.Action{
			Type:    browserk.ActLeftClick,
			Element: ele,
		}

		_, causedLoad, err := b.ExecuteAction(ctx, act)
		if err != nil {
			t.Fatalf("error executing click: %s\n", err)
		}

		if causedLoad {
			t.Fatalf("load should not have been caused")
		}
	}

	evts := b.GetConsoleEvents()
	if evts == nil {
		t.Fatalf("console events were not captured")
	}
	if len(evts) != 3 {
		t.Fatalf("expected 3 console log events, got %d\n", len(evts))
	}
}
