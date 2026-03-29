# TD3Patch

Test Drive 3 runs much too fast on almost any modern computer,
while any computer slow enough to run the game at or near its
intended frame rate suffers from excessive variability and slowdown
when there is a lot happening on screen.

However, while testing on a fast Pentium I noticed that it never
exceeds the VGA vertical refresh. Armed with this suspicion, I
investigated the code where TD3 waits for the retrace signal and
realised there are enough bytes spare to wait multiple times,
thus slowing the framerate.

This does remove two memory checks which might have been essential
for something, so I cannot guarantee this doesn't break something
on weird hardware configurations. It works for both VGA and EGA
modes, although be aware of potential differing refresh rates and
therefore different delay values being needed for each.

It's also likely this does not handle all available executables of
Test Drive 3 that can be found in the wild. I've tested it against
a few commonly available abandonware releases including Quader's
so if all else fails try that.

## Building

`go build main/td3patch.go`

## Usage

`td3patch PATH/TO/TDIII.EXE`
`td3patch -delay x PATH/TO/TD3.EXE`

The `delay` parameter allows you to choose how many vblank cycles
to wait before proceeding to the next frame. 6 appears to be a
playable default but still results in time running too fast; the
original design for TD3 appears to have been intended to run at
a much lower frame rate.

TD3patch will correctly handle patching an already-patched file
if you wish to play with frame rates to find a suitable one.