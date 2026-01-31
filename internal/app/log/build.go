package log

import infralog "workforge/internal/infra/log"

func Verbose() bool {
	return infralog.Verbose()
}
