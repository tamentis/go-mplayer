// Copyright 2014, Bertrand Janin <b@janin.com>. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// All the commands available to an MPlayer slave process are available in the
// MPlayer docs folder, also available online:
//
//     http://www.mplayerhq.hu/DOCS/tech/slave.txt
//

package mplayer

import (
	"bufio"
	"io"
	"os/exec"
	"strings"
	"time"
)

var (
	// Input is used to feed the slave subprocess commands.
	Input = make(chan string)

	// MPlayer has stopped playing.
	stoppedCh = make(chan bool)

	// hasStopSignalListeners signals that there are listened for the Stop
	// signal emitted
	hasStopSignalListeners = false
)

// readOutput is a go routine transferring the stdout of MPlayer to a proper
// channel.
func readOutput(reader io.Reader) {
	bufReader := bufio.NewReader(reader)

	for {
		msg, err := bufReader.ReadString('\n')
		if err != nil {
			// process likely died, let the routine die as well
			return
		}
		msg = strings.TrimSpace(msg)

		// When PlayAndWait is used, we send a "get_property path" to
		// MPlayer every seconds.  If the response is ever that nothing
		// is player anymore, we shutdown the player.
		if hasStopSignalListeners && msg == "ANS_path=(null)" {
			stoppedCh <- true
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

	err = cmd.Start()
	if err != nil {
		return err
	}

	go readOutput(reader)

	for msg := range Input {
		_, err = writer.Write([]byte(msg + "\n"))
		if err != nil {
			break
		}
	}

	cmd.Wait()

	return err
}

// keepSlaveAlive is a loop keeping at least one instance of MPlayer going
// continuously.  If a process exists a new one is created, possibly with a
// delay if the previous one exited with an error.  If you want some reporting
// to happen, you need to define an error handler.
func keepSlaveAlive(errorHandler func(...interface{})) {
	for {
		err := runProcess()

		if hasStopSignalListeners {
			stoppedCh <- true
		}

		if err != nil {
			if errorHandler != nil {
				errorHandler(err)
			}
			time.Sleep(10 * time.Second)
		}
	}
}

// StartSlave keeps an mplayer process in slave mode open for the rest of time,
// consuming input from the mplayerInput channel and feeding it to the stdin of
// the process. If the process dies it is restarted automatically.
//
// You are required to define an error handler function that will be called
// with all the errors that could have occured managing the slave.
func StartSlave(errorHandler func(...interface{})) {
	go keepSlaveAlive(errorHandler)
}

// SendCommand feeds the MPlayer slave with input commands.
func SendCommand(msg string) {
	Input <- msg
}
