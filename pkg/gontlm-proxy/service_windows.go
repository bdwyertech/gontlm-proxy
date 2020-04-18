// +build windows

// Windows Service Handler

package ntlm_proxy

import (
	"flag"
	log "github.com/sirupsen/logrus"

	"github.com/kardianos/service"
	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
	"os"
)

type program struct {
	exit chan struct{}
}

var svcFlag string

func init() {
	if flag.Lookup("service") == nil {
		flag.StringVar(&svcFlag, "service", "", "Control the Windows System service.")
	}
}

func RunWindows() {
	svcConfig := &service.Config{
		Name:        "gontlm-proxy",
		DisplayName: "GoNTLM Proxy",
		Description: "GoNTLM Forwarding Proxy",
		Arguments:   []string{"-proxy", proxyServer},
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}

	if len(svcFlag) != 0 {
		err := service.Control(s, svcFlag)
		if err != nil {
			log.Printf("Valid actions: %q\n", service.ControlAction)
			log.Fatal(err)
		}
		return
	}

	if err = s.Run(); err != nil {
		log.Fatal(err)
	}
}

func (p *program) Start(s service.Service) (err error) {
	if service.Interactive() {
		if isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) {
			log.SetFormatter(&log.TextFormatter{ForceColors: true})
			log.SetOutput(colorable.NewColorableStdout())
		}
		log.Info("Running in terminal.")
	} else {
		log.Info("Running under service manager.")
	}
	p.exit = make(chan struct{})

	go p.run()

	return
}

func (p *program) run() (err error) {
	// Run the Proxy
	go Run()
	// Wait for Exit Signal
	for {
		select {
		case <-p.exit:
			return
		}
	}
}

func (p *program) Stop(s service.Service) (err error) {
	close(p.exit)
	return
}
