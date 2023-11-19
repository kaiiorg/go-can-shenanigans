package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"go.einride.tech/can"
	"go.einride.tech/can/pkg/socketcan"
)

var (
	logLevel = flag.String("log-level", "info", "Zerolog log level")
	canIp    = flag.String("ip", "", "IP address of CAN device")
	canPort  = flag.Int("port", 0, "Port of CAN device")
)

func main() {
	flag.Parse()
	configureLogging()
	validateFlags()

	canRecv, canRecvCloseOnceFunc := setupCanReceiver()
	defer canRecvCloseOnceFunc()

	wg, frameChan, errFrameChan := setupReceiveFrames(canRecv)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-sigChan:
			log.Info().Msg("Got signal; closing and waiting for go routines to exit")
			canRecvCloseOnceFunc()
			waitOrTimeout(wg, 3*time.Second)
			log.Info().Msg("Exiting")
			return
		case frame := <-frameChan:
			log.Info().Str("canFrame", frame.String()).Msg("Received CAN frame")
		case errFrame := <-errFrameChan:
			log.Warn().Str("errorFrame", errFrame.String()).Msg("Received CAN error frame")
		}
	}
}

func connectToCanDevice() net.Conn {
	log.Debug().Str("ip", *canIp).Int("port", *canPort).Msg("Attempting to connect to CAN device over TCP")
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", *canIp, *canPort))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect via TCP to CAN device")
	}
	log.Debug().Msg("TCP ok")
	return conn
}

func setupCanReceiver() (*socketcan.Receiver, func()) {
	canRecv := socketcan.NewReceiver(connectToCanDevice())
	// Use sync.OnceFunc to make sure we call canRecv.Close() exactly once, so we can defer calling it.
	// This is just to make sure the underlying TCP connection gets closed cleanly should we explode
	canRecvCloseOnceFunc := sync.OnceFunc(
		func() {
			log.Warn().Msg("Closing CAN receiver")
			canRecv.Close()
		},
	)
	return canRecv, canRecvCloseOnceFunc
}

func setupReceiveFrames(canRecv *socketcan.Receiver) (*sync.WaitGroup, chan can.Frame, chan socketcan.ErrorFrame) {
	frameChan := make(chan can.Frame, 10)
	errFrameChan := make(chan socketcan.ErrorFrame, 10)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go receiveFrames(wg, canRecv, frameChan, errFrameChan)
	return wg, frameChan, errFrameChan
}

func receiveFrames(wg *sync.WaitGroup, canRecv *socketcan.Receiver, frameChan chan can.Frame, errFrameChan chan socketcan.ErrorFrame) {
	log.Info().Msg("Waiting to receive CAN frames")
	for canRecv.Receive() {
		frame := canRecv.Frame()
		log.Debug().
			Uint32("ID", frame.ID).
			Uint8("Length", frame.Length).
			Bool("IsRemote", frame.IsRemote).
			Bool("IsExtended", frame.IsExtended).
			Interface("data", frame.Data).
			AnErr("ReceiverErr", canRecv.Err()).
			Bool("ReceiverHasErrorFrame", canRecv.HasErrorFrame()).
			Msg("Received CAN frame")

		if canRecv.HasErrorFrame() {
			errFrameChan <- canRecv.ErrorFrame()
			continue
		}

		frame.Length = uint8(len(frame.Data))
		frameChan <- frame
	}
	log.Debug().Msg("Go routine receiving CAN frames exiting!")
	wg.Done()
}
