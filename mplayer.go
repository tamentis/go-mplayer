// Copyright 2014, Bertrand Janin <b@janin.com>. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package mplayer

import (
	"bufio"
	"os/exec"
	"time"
	"strings"
)

const (
	RestartTimeout = 10 * time.Second
)

var (
	// Channel used to feed the subprocess.
	Input = make(chan string)

	// MPlayer has stopped playing.
	Stopped = make(chan bool)

	// If this is true, someone is listening to the Stopped channel.
	EmitStopSignal = false
)

// Type of the function being passed to the StartSlave as error handler.
type ErrorHandler func(error)

// Routine transferring the stdout of MPlayer to a proper channel.
func handleMplayerOutput(bufReader *bufio.Reader) {
	for {
		msg, err := bufReader.ReadString('\n')
		if err != nil {
			// process likely died, let the routine die as well
			return
		}
		msg = strings.TrimSpace(msg)

		// This is quite hackish but works alright for now. When
		// PlayAndWait is used, we send a "get_property path" to
		// MPlayer every seconds. If the response is ever that nothing
		// is player anymore, we shutdown the player.
		if EmitStopSignal && msg == "ANS_path=(null)" {
			Stopped <- true
		}
	}
}

// Run a single slave process and wait for its completion before returning. If
// the process starts properly, this function will pass to its stdin all the
// message on the Input channel.
func runProcess() error {
	cmd := exec.Command("mplayer", "-quiet", "-slave", "-idle")
	writer, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	reader, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	bufReader := bufio.NewReader(reader)

	err = cmd.Start()
	if err != nil {
		return err
	}

	go handleMplayerOutput(bufReader)

	for msg := range Input {
		_, err = writer.Write([]byte(msg + "\n"))
		if err != nil {
			break
		}
	}

	cmd.Wait()

	return err
}

// Continuously restart MPlayer. If any error occurs, wait a moment before
// restarting.
func keepSlaveAlive(errorHandler ErrorHandler) {
	for {
		err := runProcess()

		if EmitStopSignal {
			Stopped <- true
		}

		if err != nil {
			errorHandler(err)
			time.Sleep(RestartTimeout)
		}
	}
}

// Keep an mplayer process in slave mode open for the rest of time, consuming
// input from the mplayerInput channel and feeding it to the stdin of the
// process. If the process dies it is restarted automatically.
//
// You are required to define an error handler function that will be called
// with all the errors that could have occured managing the slave.
func StartSlave(errorHandler ErrorHandler) {
	go keepSlaveAlive(errorHandler)
}

// Feed the MPlayer slave with input commands.
func SendCommand(msg string) {
	Input <- msg
}

// Play the given file and block until the file is done playing.
func PlayAndWait(path string) {
	SendCommand("loadfile "+path)
	EmitStopSignal = true

	// Send a query for the path every seconds. The response is expected in
	// handleMplayerOutput.
	ticker := time.Tick(time.Second)

	for {
		select {
		case <-Stopped:
			EmitStopSignal = false
			return
		case <-ticker:
			SendCommand("get_property path")
		}
	}
}

// Play the given file and block until the file is done playing. This function
// will also stop playing after the given duration.
func PlayAndWaitWithDuration(path string, duration time.Duration) {
	go func() {
		time.Sleep(duration)
		SendCommand("stop")
	}()

	PlayAndWait(path)
}
