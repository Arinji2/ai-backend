package custom_log

import (
	"github.com/gookit/slog"
	"github.com/gookit/slog/handler"
)

var Logger *slog.Logger

func init() {

	template := `[{{level}}] [{{caller}}] {{message}}`

	h := handler.NewConsoleHandler(slog.AllLevels)
	h.Formatter().(*slog.TextFormatter).SetTemplate(template)

	Logger = slog.NewWithHandlers(h)
}
