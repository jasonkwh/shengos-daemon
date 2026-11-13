package clipboard

import (
	"github.com/jasonkwh/shengos-daemon/internal/message"
	"golang.design/x/clipboard"
)

func toMessageClipboardFormat(
	format clipboard.Format,
) message.ClipboardFormat {
	switch format {
	case clipboard.FmtText:
		return message.ClipboardFormatText
	case clipboard.FmtImage:
		return message.ClipboardFormatImage
	case clipboard.FmtFiles:
		return message.ClipboardFormatFiles
	default:
		return message.ClipboardFormatUnknown
	}
}

func toClipboardFormat(
	format message.ClipboardFormat,
) clipboard.Format {
	switch format {
	case message.ClipboardFormatText:
		return clipboard.FmtText
	case message.ClipboardFormatImage:
		return clipboard.FmtImage
	case message.ClipboardFormatFiles:
		return clipboard.FmtFiles
	default:
		return clipboard.FmtText
	}
}
