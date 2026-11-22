# Using homewire

How the hub keeps its devices, and how devices are added, shared and removed. See the [README](../README.md) for installing it.

## Fabric and nodes

The hub is a Matter controller with a fabric of its own. It does not use the store of `matterctl` or of another go-matter application, so a device commissioned with those is not a device of the hub.

- **First start.** The hub creates its fabric in the data directory (`fabric.json`) and logs its ID. Later starts use that fabric; it is never replaced. A data directory which holds commissioned nodes but no fabric is refused.
- **Nodes.** At every start the hub connects to each node commissioned into its fabric (`commissions/` in the data directory) and subscribes to all its attributes. It keeps the last value of every attribute in memory.
- **Status.** A node is `connecting` until its first report, `online` while reports arrive, and `offline` when it cannot be reached or its subscription is lost. The hub keeps trying, waiting from 2 seconds up to 2 minutes between attempts, and keeps the last values meanwhile. It reconnects as soon as the device can be reached again. After the Wi-Fi network drops, some devices do not get back on it by themselves, or stay half-connected, sending but not receiving: turn such a device off and on again.
- **Log.** A change of status and every changed attribute is logged:

```
level=INFO msg="node status" node=0000000016A5E84A status=online
level=INFO msg="attribute changed" node=0000000016A5E84A endpoint=1 cluster=0x0006 attribute=0x0000 value=true
```

`fabric.json` holds the private keys of the fabric, unencrypted, readable by the hub's user only (mode 0600). Whoever can read it, root included, can control every device of the hub. Under the system service, that user is `homewire`, which nothing else runs as. A key store, which would keep the keys encrypted, is not implemented.

## Adding a device

**Add device** in the web UI takes the device's setup code: the text of its QR code (`MT:...`) or the 11- or 21-digit manual pairing code printed beside it.

<img src="docs/images/add-device.png" alt="Add device: the setup code, and the Wi-Fi network for a new device" width="380">

- **A new device**, not on any network yet, is found over Bluetooth and given the Wi-Fi network it is to join. The host of the hub needs a Bluetooth adapter (BlueZ on Linux), and most devices join a 2.4 GHz network only. The browser remembers the network name; the password is handed to the device and not stored.
- **A device already on the network**, Wi-Fi or Ethernet, is found by mDNS. It has to be waiting to be set up: new from the factory or reset, or with a pairing window opened from another app or hub it is set up with.

The hub adds one device at a time, for up to three minutes. Once commissioned, the device joins the hub's list at once and is followed like the others. A device that is on the network but not waiting to be set up (it advertises `CM=0`) is reported right away instead of after a timeout. If another app or hub has it, open a pairing window there; otherwise reset it to factory defaults. A device is left that way after it was removed from this hub, and by some devices after a commissioning which failed midway, once they have joined the Wi-Fi network.

Neither the setup code nor the Wi-Fi password is logged or reported back by the API, at any log level: the hex dumps go-matter makes of the messages it sends and receives, which may hold them, are never logged.

## Sharing a device with another app

A device of the hub can be added to another app or hub as well, such as Google Home, Amazon Alexa, Apple Home or SmartThings, and is then controlled from both; the hub stays in control (Matter multi-admin).

1. **Share with another app…** on the device's card opens a pairing window on the device for 10 minutes, and shows a QR code and an 11-digit setup code for it.
2. In the other app, add a Matter device, and scan the QR code or type the setup code. The code works once, while the window is open.
3. The device then belongs to both. **Close the window** ends sharing early.

<img src="docs/images/share.png" alt="Sharing a device: the QR code and the setup code of its pairing window" width="380">

The other way round works too: in the other app, open a pairing window for a device it has (in Google Home, *Linked Matter apps & services*), and add it to the hub with **Add device** and that code. The device is found on the network.

The hub does not log or keep the codes: they are made for one window, and shown once.

## Removing a device

**Remove device…** under a device's attributes removes it, after a confirmation. The hub sends the device RemoveFabric over its session with it, so the device itself leaves the hub's fabric; then the hub stops following it and deletes its record (`commissions/` in the data directory).

A device which is offline, or which does not leave the fabric, is not removed. The hub can then forget it anyway (**Remove from the hub only**): the device keeps the hub's fabric until it is reset to factory defaults, which it needs before it can be added again.

To add a removed device again, open a pairing window for it in another app or hub which still has it, and add it with that code over the network; a reset is not needed then. Otherwise reset it to factory defaults: a device may start waiting to be set up again by itself once it has no fabric left, but many do not: they stay on the network, not waiting to be set up.
