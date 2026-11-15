#!/bin/sh
set -e

# The service runs as the system user homewire.
if [ "$1" = "configure" ]; then
	if ! getent passwd homewire >/dev/null; then
		adduser --system --group --home /var/lib/homewire --no-create-home \
			--gecos "homewire Matter hub" homewire >/dev/null
	fi
	# BlueZ lets the members of the bluetooth group use the adapter.
	if getent group bluetooth >/dev/null; then
		adduser homewire bluetooth >/dev/null 2>&1 || true
	fi
fi

# The service is enabled and started at the first installation and restarted
# at an upgrade, the way debhelper does it; a service the administrator
# disabled stays disabled.
case "$1" in
configure | abort-upgrade | abort-deconfigure | abort-remove)
	deb-systemd-helper unmask homewire.service >/dev/null || true
	if deb-systemd-helper --quiet was-enabled homewire.service; then
		deb-systemd-helper enable homewire.service >/dev/null || true
	else
		deb-systemd-helper update-state homewire.service >/dev/null || true
	fi
	if [ -d /run/systemd/system ]; then
		systemctl --system daemon-reload >/dev/null || true
		if [ -n "$2" ]; then
			deb-systemd-invoke restart homewire.service >/dev/null || true
		else
			deb-systemd-invoke start homewire.service >/dev/null || true
		fi
	fi
	;;
esac
