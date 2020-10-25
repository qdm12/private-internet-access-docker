package logging

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/qdm12/gluetun/internal/constants"
	"github.com/qdm12/golibs/logging"
)

//nolint:lll
var regularExpressions = struct { //nolint:gochecknoglobals
	unboundPrefix     *regexp.Regexp
	tinyproxyLoglevel *regexp.Regexp
	tinyproxyPrefix   *regexp.Regexp
}{
	unboundPrefix:     regexp.MustCompile(`unbound: \[[0-9]{10}\] unbound\[[0-9]+:0\] `),
	tinyproxyLoglevel: regexp.MustCompile(`INFO|CONNECT|NOTICE|WARNING|ERROR|CRITICAL`),
	tinyproxyPrefix:   regexp.MustCompile(`tinyproxy: .+[ ]+(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec) [0-3][0-9] [0-2][0-9]:[0-5][0-9]:[0-5][0-9] \[[0-9]+\]: `),
}

func PostProcessLine(s string) (filtered string, level logging.Level) {
	switch {
	case strings.HasPrefix(s, "openvpn: "):
		for _, ignored := range []string{
			"openvpn: WARNING: you are using user/group/chroot/setcon without persist-tun -- this may cause restarts to fail",
			"openvpn: NOTE: UID/GID downgrade will be delayed because of --client, --pull, or --up-delay",
		} {
			if s == ignored {
				return "", ""
			}
		}
		switch {
		case strings.HasPrefix(s, "openvpn: NOTE: "):
			filtered = strings.TrimPrefix(s, "openvpn: NOTE: ")
			filtered = "openvpn: " + filtered
			level = logging.InfoLevel
		case strings.HasPrefix(s, "openvpn: WARNING: "):
			filtered = strings.TrimPrefix(s, "openvpn: WARNING: ")
			filtered = "openvpn: " + filtered
			level = logging.WarnLevel
		case strings.HasPrefix(s, "openvpn: Options error: "):
			filtered = strings.TrimPrefix(s, "openvpn: Options error: ")
			filtered = "openvpn: " + filtered
			level = logging.ErrorLevel
		case s == "openvpn: Initialization Sequence Completed":
			return color.HiGreenString(s), logging.InfoLevel
		case s == "openvpn: AUTH: Received control message: AUTH_FAILED":
			filtered = s + "\n\n  (IF YOU ARE USING PIA V4 servers, MAYBE CHECK OUT https://github.com/qdm12/gluetun/issues/265)\n" //nolint:lll
			level = logging.ErrorLevel
		default:
			filtered = s
			level = logging.InfoLevel
		}
		filtered = constants.ColorOpenvpn().Sprintf(filtered)
		return filtered, level
	case strings.HasPrefix(s, "unbound: "):
		prefix := regularExpressions.unboundPrefix.FindString(s)
		filtered = s[len(prefix):]
		switch {
		case strings.HasPrefix(filtered, "notice: "):
			filtered = strings.TrimPrefix(filtered, "notice: ")
			level = logging.InfoLevel
		case strings.HasPrefix(filtered, "info: "):
			filtered = strings.TrimPrefix(filtered, "info: ")
			level = logging.InfoLevel
		case strings.HasPrefix(filtered, "warn: "):
			filtered = strings.TrimPrefix(filtered, "warn: ")
			level = logging.WarnLevel
		case strings.HasPrefix(filtered, "error: "):
			filtered = strings.TrimPrefix(filtered, "error: ")
			level = logging.ErrorLevel
		default:
			level = logging.ErrorLevel
		}
		filtered = fmt.Sprintf("unbound: %s", filtered)
		filtered = constants.ColorUnbound().Sprintf(filtered)
		return filtered, level
	case strings.HasPrefix(s, "tinyproxy: "):
		logLevel := regularExpressions.tinyproxyLoglevel.FindString(s)
		prefix := regularExpressions.tinyproxyPrefix.FindString(s)
		filtered = fmt.Sprintf("tinyproxy: %s", s[len(prefix):])
		filtered = constants.ColorTinyproxy().Sprintf(filtered)
		switch logLevel {
		case "INFO", "CONNECT", "NOTICE":
			return filtered, logging.InfoLevel
		case "WARNING":
			return filtered, logging.WarnLevel
		case "ERROR", "CRITICAL":
			return filtered, logging.ErrorLevel
		default:
			return filtered, logging.ErrorLevel
		}
	}
	return s, logging.InfoLevel
}
