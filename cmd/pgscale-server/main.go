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

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"

	"github.com/buraksezer/olric"
	"github.com/pgscale/pgscale/cmd/pgscale-server/server"
	"github.com/sean-/seed"
)

func usage() {
	var msg = `Usage: pgscale-server [options] ...

Distributed query cache and connection pool middleware for PostgreSQL.

Options:
  -h, --help       Print this message and exit.
  -v, --version    Print the version number and exit.
  -c, --config     Set configuration file path.

The Go runtime version %s
Report bugs to https://github.com/pgscale/pgscale/issues
`
	_, err := fmt.Fprintf(os.Stdout, msg, runtime.Version())
	if err != nil {
		panic(err)
	}
}

type arguments struct {
	config  string
	help    bool
	version bool
}

var (
	Version string
	GitSHA  string
)

const (
	// EnvConfigFile is the name of environment variable which can be used to override default configuration file path.
	EnvConfigFile     = "PGSCALE_SERVER_CONFIG"
	DefaultConfigFile = "pgscale-server.hcl"
)

func main() {
	args := &arguments{}
	// No need for timestamp etc in this function. Just log it.
	log.SetFlags(0)

	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("pgscale-server: failed to find current directory: %v", err)
	}
	configFile := path.Join(currentDir, DefaultConfigFile)

	// Parse command line parameters
	f := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	f.SetOutput(ioutil.Discard)
	f.BoolVar(&args.help, "h", false, "")
	f.BoolVar(&args.help, "help", false, "")

	f.BoolVar(&args.version, "version", false, "")
	f.BoolVar(&args.version, "v", false, "")

	f.StringVar(&args.config, "config", configFile, "")
	f.StringVar(&args.config, "c", configFile, "")

	if err := f.Parse(os.Args[1:]); err != nil {
		log.Fatalf("pgscale-server: failed to parse flags: %v", err)
	}

	if args.version {
		fmt.Printf("PgScale Version: %s\n", Version)
		fmt.Printf("Git SHA: %s\n", GitSHA)
		fmt.Printf("Olric Version: %s\n", olric.ReleaseVersion)
		fmt.Printf("Go Version: %s\n", runtime.Version())
		fmt.Printf("Go OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		return
	} else if args.help {
		usage()
		return
	}

	// MustInit provides guaranteed secure seeding.  If `/dev/urandom` is not
	// available, MustInit will panic() with an error indicating why reading from
	// `/dev/urandom` failed.  MustInit() will upgrade the seed if for some reason a
	// call to Init() failed in the past.
	seed.MustInit()

	envConfigFile := os.Getenv(EnvConfigFile)
	if envConfigFile != "" {
		args.config = envConfigFile
	}

	s, err := server.New(args.config)
	if err != nil {
		log.Fatalf("pgscale-server: %v", err)
	}

	if err = s.Start(); err != nil {
		log.Fatalf("pgscale-server: %v", err)
	}

	log.Print("Quit!")
}
