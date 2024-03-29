package browser

import (
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/wirepair/gcd"
)

// LocalLeaser for leasing locally
type LocalLeaser struct {
	browserLock    sync.RWMutex
	browsers       map[string]*gcd.Gcd
	browserTimeout time.Duration
	tmp            string
	chromeLocation string
}

// NewLocalLeaser for browsers
func NewLocalLeaser() *LocalLeaser {
	s := &LocalLeaser{
		browserLock:    sync.RWMutex{},
		browserTimeout: time.Second * 30,
		browsers:       make(map[string]*gcd.Gcd),
	}
	s.chromeLocation, s.tmp = FindChrome()
	log.Info().Msgf("FOUND CHROME %s and TMP: %s", s.chromeLocation, s.tmp)
	return s
}

// Acquire a new browser
func (s *LocalLeaser) Acquire() (string, error) {
	b := gcd.NewChromeDebugger()
	b.DeleteProfileOnExit()

	profileDir := randProfile(s.tmp)
	port := randPort()
	log.Info().Msgf("chrome temp %s path: %s", s.tmp, profileDir)
	b.AddFlags(startupFlags)
	if err := b.StartProcess(s.chromeLocation, profileDir, port); err != nil {
		return "", err
	}
	s.browserLock.Lock()
	s.browsers[port] = b
	s.browserLock.Unlock()

	return string(port), nil
}

// Count how many browsers
func (s *LocalLeaser) Count() (string, error) {
	s.browserLock.RLock()
	count := len(s.browsers)
	s.browserLock.RUnlock()
	return strconv.Itoa(count), nil
}

// Return (and kill) the browser
func (s *LocalLeaser) Return(port string) error {
	s.browserLock.Lock()
	defer s.browserLock.Unlock()

	if b, ok := s.browsers[port]; ok {
		if err := b.ExitProcess(); err != nil {
			return err
		}
		delete(s.browsers, port)
		return nil
	}

	return errors.New("not found")
}

// Cleanup all old browser processes, hope you weren't running chrome!
func (s *LocalLeaser) Cleanup() (string, error) {
	if err := KillOldProcesses(); err != nil {
		return "", err
	}

	if s.tmp == "" {
		log.Fatal().Msg("tmp directory is empty! this could have deleted system files, exiting")
	}

	if err := RemoveTmpContents(s.tmp); err != nil {
		return "", err
	}
	return "ok", nil
}
