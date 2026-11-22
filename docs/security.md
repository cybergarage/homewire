# Access control and signing in

Who can reach the hub, and how a browser or a script signs in.

## Access control

By default the hub listens on the loopback interface only, and needs no sign-in: only programs on the same host can reach it.

To use it from other machines, listen on a network address:

```sh
bin/homewire serve -listen 0.0.0.0:8580
```

The hub then always requires its **access token**. This cannot be turned off.

- The token is created at the first start, 256 random bits, in `<data-dir>/token` with mode 0600. The hub logs the path of the file, never the token. A token file other users can read, or its group can write, is refused.
- In a data directory its group can read and others cannot, such as the system service's `/var/lib/homewire` (mode 0750, group `homewire`), the token and `serve.json` are created with mode 0640: the members of that group can use `homewire open`, `status` and `token`. `fabric.json` stays readable by the hub's user only.
- **Browser**: run `homewire open`, or enter the token once on the sign-in page. The hub sets a session cookie (`HttpOnly`, `SameSite=Strict`) valid for 30 days. The hub signs the session with a key derived from the token instead of keeping it, so a session outlasts a restart of the hub. Signing out refuses that session until the hub restarts.
- **Scripts**: send `Authorization: Bearer <token>`.
- After 10 wrong tokens within a minute, sign-in is refused for the rest of that minute.
- To replace the token, delete the file and restart the hub. That also ends every browser session.

```sh
curl -H "Authorization: Bearer $(cat ~/.homewire/token)" http://192.168.1.20:8580/api/v1/hub
```

Two checks apply whether or not the token is required. They stop a web page on another site from using the browser of someone on the home network to reach the hub:

- **Host header.** A request has to address the hub by an IP address, by `localhost`, by the host name of the machine (also with `.local`), or by a name given with `-allowed-host`. Other names are refused, which defeats DNS rebinding.
- **Origin.** A request which changes something has to come from a page the hub itself served, and a request body has to be JSON.

## What this does not cover

The hub speaks plain HTTP. On the network, the token (at sign-in and in a bearer header) and the session cookie are not encrypted, so anyone who can capture traffic between the browser and the hub can read them. Use it on a network you trust, or put it behind a TLS reverse proxy. TLS in the hub itself is not implemented.

## Signing a browser in

`homewire open`, run on the host of the hub as the user who runs it or, for the system service, as a member of the group `homewire`, opens the web UI in the default browser, already signed in: nobody types or copies the token.

```sh
homewire open                                  # a browser on this host
homewire open -print                           # print the link instead, for a browser on another machine
ssh hub.local homewire open -print             # the same, from that other machine
homewire open -print -host 192.168.1.20        # name the hub by another name or address in the link
```

It reads the token from the data directory, asks the hub for a **one-time sign-in code**, and opens `/signin?code=...`. The hub exchanges the code for a session and sends the browser on to the web UI, so the code leaves the address bar. A code works once, within a minute; a link used again, or too late, leads to the sign-in page.

With `-print`, the link names the hub by this host's `.local` name when the hub listens on the network; `-host` names it otherwise.
