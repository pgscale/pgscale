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

package dante

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"syscall"

	"github.com/buraksezer/dante/config"
	"github.com/buraksezer/dante/dmaps"
	"github.com/buraksezer/dante/kontext"
	"github.com/buraksezer/dante/postgresql"
	"github.com/buraksezer/dante/utils"
	"github.com/buraksezer/olric"
	"github.com/buraksezer/olric/pkg/flog"
	"github.com/hashicorp/logutils"
)

type Dante struct {
	config            *config.Config
	log               *flog.Logger
	postgres          *postgresql.PostgreSQL
	olric             *olric.Olric
	dmaps             *dmaps.DMaps
	shutdownCallbacks []func()
}

func (d *Dante) configureLogger(c *config.Config) (*flog.Logger, error) {
	var out *os.File

	switch c.Dante.Logging.Output {
	case config.Stderr:
		out = os.Stderr
	case config.Stdout:
		out = os.Stdout
	default:
		f, err := os.OpenFile(c.Dante.Logging.Output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, c.Dante.Logging.Perm)
		if err != nil {
			return nil, err
		}
		out = f
		d.shutdownCallbacks = append(d.shutdownCallbacks, func() {
			if err := f.Close(); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "failed to close log file: %v", err)
			}
		})
	}

	filter := &logutils.LevelFilter{
		Levels: []logutils.LogLevel{
			config.DebugLog,
			config.WarnLog,
			config.ErrorLog,
			config.InfoLog,
		},
		MinLevel: logutils.LogLevel(strings.ToUpper(c.Dante.Logging.Level)),
		Writer:   out,
	}
	l := log.New(out, "", log.LstdFlags)
	l.SetOutput(filter)
	logger := flog.New(l)
	logger.SetLevel(c.Dante.Logging.Verbosity)
	if c.Dante.Logging.Level == config.DebugLog {
		logger.ShowLineNumber(1)
	}
	return logger, nil
}

func New(k *kontext.Kontext) (*Dante, error) {
	c, err := config.FromKontext(k)
	if err != nil {
		return nil, err
	}

	db, err := utils.OlricFromKontext(k)
	if err != nil {
		return nil, err
	}

	d := &Dante{
		config:            c,
		olric:             db,
		dmaps:             dmaps.New(db),
		shutdownCallbacks: []func(){},
	}
	l, err := d.configureLogger(c)
	if err != nil {
		return nil, err
	}
	d.log = l
	return d, err
}

func (d *Dante) ListenAndServe() error {
	k := kontext.New()
	k.Set(kontext.LoggerKey, d.log)
	k.Set(kontext.ConfigKey, d.config)
	k.Set(kontext.DMapsKey, d.dmaps)
	p, err := postgresql.New(k)
	if err != nil {
		return err
	}
	d.postgres = p

	err = p.ListenAndServe()
	opErr, ok := err.(*net.OpError)
	if !ok {
		return err
	}

	switch {
	case opErr.Op == "accept" && opErr.Err.Error() == "use of closed network connection":
		return nil
	case errors.Is(opErr.Err, syscall.EPIPE) || opErr.Op == "read":
		// write: broken pipe
		// read: use of closed network connection
		d.log.V(3).Printf("[DEBUG] TCP server has returned an error: %v", err)
		return nil
	case errors.Is(opErr.Err, syscall.ECONNRESET):
		// read: connection reset by peer
		d.log.V(3).Printf("[DEBUG] TCP server has returned an error: %v", err)
		return nil
	default:
		return err
	}
}

func (d *Dante) Shutdown() error {
	if d.shutdownCallbacks != nil {
		for _, f := range d.shutdownCallbacks {
			f()
		}
	}
	return d.postgres.Shutdown()
}
