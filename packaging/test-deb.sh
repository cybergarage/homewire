#!/bin/sh
# Installs a homewire .deb on this host and checks the service it sets up:
# run by CI on a disposable runner, with sudo.
#
#   packaging/test-deb.sh dist/homewire_<version>_amd64.deb
set -eu

deb=$1
url=http://127.0.0.1:8580

fail() {
	echo "test-deb: $*" >&2
	sudo journalctl -u homewire --no-pager -n 50 >&2 || true
	exit 1
}

wait_healthy() {
	for _ in $(seq 1 20); do
		if curl -fsS "$url/api/v1/health" >/dev/null 2>&1; then
			return 0
		fi
		sleep 1
	done
	fail "the hub does not answer on $url"
}

echo "== install"
sudo apt-get install -y "./$deb"
systemctl is-enabled --quiet homewire || fail "the service is not enabled"
systemctl is-active --quiet homewire || fail "the service is not running"
wait_healthy
curl -fsS "$url/api/v1/health"
echo

echo "== data directory"
sudo sh -c "stat -c '%U:%G %a %n' /var/lib/homewire /var/lib/homewire/*"
[ "$(sudo stat -c '%U:%G %a' /var/lib/homewire)" = "homewire:homewire 750" ] || fail "unexpected owner or mode of /var/lib/homewire"
[ "$(sudo stat -c '%a' /var/lib/homewire/fabric.json)" = "600" ] || fail "fabric.json is not mode 0600"
[ "$(sudo stat -c '%a' /var/lib/homewire/serve.json)" = "640" ] || fail "serve.json is not mode 0640"

echo "== launcher and icon"
test -f /usr/share/applications/homewire.desktop || fail "no launcher in /usr/share/applications"
grep -qx 'Exec=homewire open -notify' /usr/share/applications/homewire.desktop || fail "the launcher does not run homewire open -notify"
test -f /usr/share/icons/hicolor/scalable/apps/homewire.svg || fail "no scalable icon"
test -f /usr/share/icons/hicolor/48x48/apps/homewire.png || fail "no 48x48 icon"
if command -v desktop-file-validate >/dev/null; then desktop-file-validate /usr/share/applications/homewire.desktop || fail "the launcher is not valid"; fi

echo "== homewire status by a member of the homewire group"
sudo usermod -a -G homewire "$(id -un)"
sg homewire -c "homewire status" || fail "homewire status failed"

echo "== reinstall"
sudo apt-get install -y --reinstall "./$deb"
systemctl is-active --quiet homewire || fail "the service is not running after a reinstall"
wait_healthy

echo "== remove"
sudo apt-get remove -y homewire
if systemctl is-active --quiet homewire; then
	fail "the service still runs after the package was removed"
fi
sudo test -f /var/lib/homewire/fabric.json || fail "remove deleted the fabric"
[ -f /etc/default/homewire ] || fail "remove deleted /etc/default/homewire"

echo "== install after remove"
sudo apt-get install -y "./$deb"
systemctl is-active --quiet homewire || fail "the service is not running after installing it again"
wait_healthy

echo "== purge"
sudo apt-get purge -y homewire
[ ! -e /var/lib/homewire ] || fail "purge left /var/lib/homewire"
[ ! -e /etc/default/homewire ] || fail "purge left /etc/default/homewire"
if getent passwd homewire >/dev/null; then
	fail "purge left the user homewire"
fi
echo "test-deb: ok"
