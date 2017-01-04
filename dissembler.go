// Copyright © 2015-2017
// Licensed under BSD 3-Clause "New" or "Revised". All rights reserved.
// Created by Christian R. Vozar <cvozar@xumak.com> in New Orleans ⚜

package dissembler

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/uber-go/zap"
)

const (
	// SIGINT is sent when a user wishes to interrupt the process; typically
	// initiated by pressing Ctrl-C, but on some systems, the "delete" character
	// or "break" key can be used.
	SIGINT = syscall.SIGINT
	// SIGHUP is sent when a user wishes to reload configuration files and reopen
	// their logfiles instead of exiting.
	SIGHUP = syscall.SIGHUP
	// SIGQUIT is sent when the user requests that the process quit and perform
	// a core dump.
	SIGQUIT = syscall.SIGQUIT
	// SIGTERM is sent to request termination. Unlike SIGKILL, it can be caught
	// and interpreted or ignored. This allows nice termination releasing
	// resources and saving state if appropriate. SIGINT is nearly identical to
	// SIGTERM.
	SIGTERM = syscall.SIGTERM
	// SIGUSR1 is sent
	SIGUSR1 = syscall.SIGUSR1
	// SIGUSR2 is sent
	SIGUSR2 = syscall.SIGUSR2
)

var (
	// Registered is the currently registered Dissembler.
	Registered Lifecycle
	// DissemblerLogger is
	DissemblerLogger log.Logger
)

// Lifecycle is the lifecycle of a represented API, service, or application.
type Lifecycle interface {
	// Init is called to do any setup of client libraries or initializing of
	// configuration prior to any operations.
	Init() error
	// Start is called to begin execution of interfacing to APIs, services, or
	// provisioning of resources.
	Start() error
	// Stop is called to perform any tear-down or deallocation of resources prior
	// to exiting.
	Stop() error
}

// Reloader is an optional interface that may be implemented by a Lifecycle to
// support Unix SIGHUP signals and reloading of configuration conditions.
//
// If a Lifecycle does not implement Reload, Dissembler will return an error
// that the Lifecycle does not support reloading of configuration.
type Reloader interface {
	Reload() error
}

// Dissembler is
type Dissembler struct {
	lifecycle Lifecycle
}

func init() {
	DissemblerLogger = log.New(
		log.NewJSONEncoder(
			log.RFC3339Formatter("timestamp"),
			log.MessageKey("message"),
			log.LevelString("level"),
		),
		log.Fields(
			log.String("dissembler_version", Version),
		),
	)
}

// Register makes a dissembler available.
//func Register(lc Lifecycle) *Dissembler {
//	return &Dissembler{Registered: lc}
//}

// Serve accepts a Dissembler lifecycle and then calls Serve with the provided
// lifecycle for the application, service, or API.
func Serve(lc Lifecycle) error {
	dissembler := &Dissembler{lifecycle: lc}
	return dissembler.Serve()
}

// Serve begins the lifecycle of the Dissembler.
func (d *Dissembler) Serve() error {
	err := d.lifecycle.Init()
	if err != nil {
		return err
	}

	/*
		err = d.lifecycle.Start()
		if err != nil {
			return err
		}*/

	// Starting process
	go func() error {
		err = d.lifecycle.Start()
		if err != nil {
			return err
		}
		return nil
	}()

	// Block and await signals
	if _, err := d.Wait(); nil != err {
		DissemblerLogger.Error("Unable to finish waiting for Dissembler to shutdown",
			log.String("error", err.Error()),
		)
	}

	return nil
}

// Wait blocks awaiting Unix signals. Signals are handled in a similar manner as
// Nginx and Unicorn: <http://unicorn.bogomips.org/SIGNALS.html>.
func (d *Dissembler) Wait() (syscall.Signal, error) {
	ch := make(chan os.Signal, 2)
	signal.Notify(
		ch,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
		syscall.SIGUSR1,
		syscall.SIGUSR2,
	)
	for {
		sig := <-ch
		DissemblerLogger.Info("signal caught",
			log.String("signal", sig.String()))
		switch sig {

		/*
					// SIGHUP reloads configuration.
			    case syscall.SIGHUP:
			    d.Registered.Reload()
			    return syscall.SIGHUP, nil
		*/

		// SIGINT should exit.
		case syscall.SIGINT:
			d.lifecycle.Stop()
			return syscall.SIGINT, nil

		// SIGQUIT should exit gracefully.
		case syscall.SIGQUIT:
			d.lifecycle.Stop()
			return syscall.SIGQUIT, nil

		// SIGTERM should exit.
		case syscall.SIGTERM:
			d.lifecycle.Stop()
			return syscall.SIGTERM, nil

			/*
				// SIGUSR2 forks and re-execs the first time it is received and execs
				// without forking from then on.
				case syscall.SIGUSR2:
					if forked {
						return syscall.SIGUSR2, nil
					}
					forked = true
					if err := ForkExec(l); nil != err {
						return syscall.SIGUSR2, err
					}
			*/
		}
	}
}
