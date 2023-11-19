package main

import (
	"context"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func configureLogging() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	logLevel, err := zerolog.ParseLevel(*logLevel)
	if err != nil || logLevel == zerolog.NoLevel {
		logLevel = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(logLevel)
}

func validateFlags() {
	if *canIp == "" {
		log.Fatal().Msg("Must provide an IP to a CAN device")
	}

	if *canPort <= 0 || *canPort > 65535 {
		log.Fatal().Int("given", *canPort).Msg("Must provide a valid CAN device port")
	}
}

func waitOrTimeout(wg *sync.WaitGroup, timeout time.Duration) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), timeout)
	go func() {
		wg.Wait()
		ctxCancel()
	}()
	<-ctx.Done()

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		log.Warn().Str("timeout", timeout.String()).Msg("Timed out while waiting for waitgroup")
	}
}
