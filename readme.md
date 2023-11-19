# Go CAN Shenanigans
This repository is me playing around with Go, [can-go](https://github.com/einride/can-go), and [WiCAN](https://github.com/meatpiHQ/wican-fw). The quality of this code should not be considered represenative of my real world I-work-40-hours-a-week output. 

I don't really have an end goal; I'm just playing around here... with the second most expensive thing I own. Maybe this is a bad idea. Note to self: _readonly_.

## Getting data from WiCAN
The documentation for [WiCAN](https://github.com/meatpiHQ/wican-fw#1-wifi) suggests using `socat` and `slcand` to make the device accessible over a local `can0` interface. 

The documentation for [can-go](https://github.com/einride/can-go#setting-up-a-can-interface) suggests using [`candevice`](https://pkg.go.dev/go.einride.tech/can/pkg/candevice) and [socketcan](https://pkg.go.dev/go.einride.tech/can/pkg/socketcan) for connecting to that local `can0` interface.

However, I noticed that [socketcan.NewReceiver](https://pkg.go.dev/go.einride.tech/can/pkg/socketcan#NewReceiver) takes an `io.ReadCloser`, which `net.Conn` just so happens to meet. I had hoped I'd be able to cut out the need for `socat`, `slcand`, and `candevice` by just `net.Dial`ing straight to the WiCAN, but this method doesn't work with my half hearted attempt at a simulator (See [simulator](#simulator)). I'm not sure if the problem is the simulator or my method of trying to pull data from it. I'll wait until I get my hardware.

## Simulator
My initial attempts to get can-go to work involved trying to use a simulator, since I made the order for the WiCAN hardware less than 8 hours ago as of writing. Long story short **this didn't work** because I don't really know what I'm doing. I'll just wait until I get my hardware.

I did this by installing can-utils and socat:
```bash
sudo apt-get install can-utils socat
```

Then, setup a `vcan` interface using the instructions [here](https://netmodule-linux.readthedocs.io/en/latest/howto/can.html#virtual-can-interface-vcan).
I wasn't able to get this to work in WSL, so I used a raspberry pi.
```bash
sudo modprobe vcan
sudo ip link add dev vcan0 type vcan
sudo ip link set vcan0 mtu 16
```

After this, I can run `cansend vcan0 5A2#11.2233.445D556677.66` in one shell and receive it with `candump vcan0` in another.

I then used `socat` to try to host that interface on a TCP port. I was lazy and asked ChatGPT for this, but it (perhaps unsurprisingly) didn't appear to work.
```bash
sudo socat -d -d TCP-LISTEN:3333,fork,reuseaddr SYSTEM:"cansend vcan0"
```

I tweaked this to use `candump` and got a bit further; my test code was able to receive _something_, though it throws an out of bounds error. Until I have hardware in hand, I'm going to assume I've set something up wrong.