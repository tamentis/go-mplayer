# go-mplayer - Go interface with the MPlayer slave-mode

Short module allowing the user to control a fork'd MPlayer process in
slave-mode.

You can read all about the MPlayer slave protocol at this address:

	http://www.mplayerhq.hu/DOCS/tech/slave.txt


## Requirements
MPlayer should be installed and available in the current PATH


## Example
This example launches MPlayer in the background, requests it to play a file,
wait for 5 seconds then stop playback:

```
import (
	"time"
	"github.com/tamentis/go-mplayer"
)

mplayer.StartSlave()

mplayer.SendCommand("loadfile /tmp/myfile.mp3")
time.Sleep(5 * time.Seconds)
mplayer.SendCommand("stop")
```

This example uses the blocking command PlayAndWait(), it allows you to use the
module in a similar fashion to `os/exec` except that you can pre-load the
process to get a better response time (e.g. on slow hardware):

```
import (
	"github.com/tamentis/go-mplayer"
)

mplayer.StartSlave()

mplayer.PlayAndWait("/tmp/myfile.mp3")
```
