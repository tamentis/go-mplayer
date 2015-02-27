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
	"time"
)

var (
	// Skip is a channel used to interrupt an existing playback.
	skipCh = make(chan bool)
)

// Attempt to cancel an existing PlayAndWait.
func Skip() {
	skipCh <- true
}

// Play the given file and block until the file is done playing.
func PlayAndWait(path string) {
	SendCommand("loadfile " + path)
	hasStopSignalListeners = true

	// Send a query for the path every seconds. The response is expected in
	// handleOutput.
	ticker := time.Tick(time.Second)

	for {
		select {
		case <-stoppedCh:
			hasStopSignalListeners = false
			return
		case <-ticker:
			SendCommand("get_property path")
		case <-skipCh:
			SendCommand("stop")
			hasStopSignalListeners = false
			return
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
