// Copyright 2021 Burak Sezer
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/buraksezer/olric"
	olricConfig "github.com/buraksezer/olric/config"
	"github.com/pgscale/pgscale/config"
	"github.com/pgscale/pgscale/kontext"
	"github.com/pgscale/pgscale/pgscale"
	"golang.org/x/sync/errgroup"
)

// Server represents a new Server instance.
type Server struct {
	log         *log.Logger
	olric       *olric.Olric
	olricConfig *olricConfig.Config
	pgscale     *pgscale.PgScale
	config      *config.Config
	errGr       errgroup.Group
	ctx         context.Context
	cancel      context.CancelFunc
}

// New creates a new Server instance
func New(configFile string) (*Server, error) {
	c, err := config.New(configFile)
	if err != nil {
		return nil, err
	}

	oc, err := config.MakeOlricConfig(c)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		log:         oc.Logger,
		olricConfig: oc,
		config:      c,
		ctx:         ctx,
		cancel:      cancel,
	}, nil
}

func (s *Server) waitForInterrupt() {
	shutDownChan := make(chan os.Signal, 1)
	signal.Notify(shutDownChan, syscall.SIGTERM, syscall.SIGINT)

	select {
	case ch := <-shutDownChan:
		s.log.Printf("[pgscale-server] Signal caught: %s", ch.String())
	case <-s.ctx.Done():
	}

	pgscaleGone := make(chan struct{})

	s.errGr.Go(func() error {
		defer close(pgscaleGone)

		return s.pgscale.Shutdown()
	})

	s.errGr.Go(func() error {
		<-pgscaleGone

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		if err := s.olric.Shutdown(ctx); err != nil {
			s.log.Printf("[pgscale-server] Failed to shutdown Olric: %v", err)
			return err
		}

		return nil
	})

	// This is not a goroutine leak. The process will quit.
	go func() {
		s.log.Printf("[pgscale-server] Awaiting for background tasks")
		s.log.Printf("[pgscale-server] Press CTRL+C or send SIGTERM/SIGINT to quit immediately")

		forceQuitCh := make(chan os.Signal, 1)
		signal.Notify(forceQuitCh, syscall.SIGTERM, syscall.SIGINT)
		ch := <-forceQuitCh

		s.log.Printf("[pgscale-server] Signal caught: %s", ch.String())
		s.log.Printf("[pgscale-server] Quits with exit code 1")
		os.Exit(1)
	}()
}

// Start starts a new PgScale server instance and blocks until the server is closed.
func (s *Server) Start() error {
	s.log.Printf("[pgscale-server] pid: %d has been started", os.Getpid())
	// Wait for SIGTERM or SIGINT
	go s.waitForInterrupt()

	ctx, cancel := context.WithCancel(context.Background())
	callback := func() {
		defer cancel()
	}
	s.olricConfig.Started = callback

	db, err := olric.New(s.olricConfig)
	if err != nil {
		return err
	}
	s.olric = db

	k := kontext.New()
	k.Set(kontext.OlricKey, db)
	k.Set(kontext.ConfigKey, s.config)

	g, err := pgscale.New(k)
	if err != nil {
		return err
	}
	s.pgscale = g

	s.errGr.Go(func() error {
		defer s.cancel()
		if err = s.olric.Start(); err != nil {
			s.log.Printf("[pgscale-server] Failed to run Olric: %v", err)
			return err
		}
		return nil
	})

	s.errGr.Go(func() error {
		defer s.cancel()
		<-ctx.Done()

		if err = s.pgscale.ListenAndServe(); err != nil {
			s.log.Printf("[pgscale-server] Failed to run PgScale: %v", err)
			return err
		}
		return nil
	})

	return s.errGr.Wait()
}
