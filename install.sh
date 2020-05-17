[ ! -d /usr/local/bin ] && echo "/usr/local/bin does not exist" && exit 1
[ ! -O /usr/local/bin ] && SUDO_MAYBE=sudo

case "$(uname -a)" in
	Linux*)  os=linux  ;;
	Darwin*) os=darwin ;;
esac

url=$(curl -Ls -o /dev/null -w %{url_effective} https://github.com/sorribas/vejr/releases/latest) 
url="${url##*/}"
$SUDO_MAYBE curl -L -o /usr/local/bin/vejr  https://github.com/sorribas/vejr/releases/download/${url}/vejr-${os}
$SUDO_MAYBE chmod +x /usr/local/bin/vejr
