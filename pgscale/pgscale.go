// Copyright 2021-2022 Burak Sezer
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

package pgscale

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"syscall"

	"github.com/buraksezer/olric"
	"github.com/buraksezer/olric/pkg/flog"
	"github.com/hashicorp/logutils"
	"github.com/pgscale/pgscale/config"
	"github.com/pgscale/pgscale/dmaps"
	"github.com/pgscale/pgscale/kontext"
	"github.com/pgscale/pgscale/postgresql"
	"github.com/pgscale/pgscale/utils"
)

type PgScale struct {
	config            *config.Config
	log               *flog.Logger
	postgres          *postgresql.PostgreSQL
	olric             *olric.Olric
	dmaps             *dmaps.DMaps
	shutdownCallbacks []func()
}

func (d *PgScale) configureLogger(c *config.Config) (*flog.Logger, error) {
	var out *os.File

	switch c.PgScale.Logging.Output {
	case config.Stderr:
		out = os.Stderr
	case config.Stdout:
		out = os.Stdout
	default:
		f, err := os.OpenFile(c.PgScale.Logging.Output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, c.PgScale.Logging.Perm)
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
		MinLevel: logutils.LogLevel(strings.ToUpper(c.PgScale.Logging.Level)),
		Writer:   out,
	}
	l := log.New(out, "", log.LstdFlags)
	l.SetOutput(filter)
	logger := flog.New(l)
	logger.SetLevel(c.PgScale.Logging.Verbosity)
	if c.PgScale.Logging.Level == config.DebugLog {
		logger.ShowLineNumber(1)
	}
	return logger, nil
}

func New(k *kontext.Kontext) (*PgScale, error) {
	c, err := config.FromKontext(k)
	if err != nil {
		return nil, err
	}

	db, err := utils.OlricFromKontext(k)
	if err != nil {
		return nil, err
	}

	d := &PgScale{
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

func (d *PgScale) ListenAndServe() error {
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

func (d *PgScale) Shutdown() error {
	if d.shutdownCallbacks != nil {
		for _, f := range d.shutdownCallbacks {
			f()
		}
	}
	return d.postgres.Shutdown()
}
