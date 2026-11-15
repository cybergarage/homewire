#!/bin/sh
set -e

if [ -d /run/systemd/system ]; then
	systemctl --system daemon-reload >/dev/null || true
fi

case "$1" in
remove)
	deb-systemd-helper mask homewire.service >/dev/null || true
	;;
purge)
	deb-systemd-helper purge homewire.service >/dev/null || true
	deb-systemd-helper unmask homewire.service >/dev/null || true
	# Purging removes the hub's data too: its fabric, its nodes and its
	# token. The devices commissioned into the fabric have to be reset to
	# factory defaults to be set up again.
	rm -rf /var/lib/homewire
	if getent passwd homewire >/dev/null; then
		deluser --system homewire >/dev/null 2>&1 || true
	fi
	if getent group homewire >/dev/null; then
		delgroup --system homewire >/dev/null 2>&1 || true
	fi
	;;
esac
