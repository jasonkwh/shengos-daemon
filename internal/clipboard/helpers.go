package clipboard

import (
	"bytes"
	"net/url"
	"strings"
)

func (c *clboard) sftpURIList(
	origin string,
	data []byte,
) ([]byte, error) {
	var body strings.Builder

	for path := range bytes.SplitSeq(data, []byte("\n")) {
		path = bytes.TrimRight(path, "\r")
		if len(path) == 0 {
			continue
		}

		if path[0] != '/' {
			return nil, errRelativePath(path)
		}

		u := url.URL{
			Scheme: "sftp",
			User:   url.User(c.cfg.SSHUser),
			Host:   origin,
			Path:   string(path),
		}

		body.WriteString(u.String())
		body.WriteString("\r\n")
	}

	if body.Len() == 0 {
		return nil, ErrEmptyPathList
	}

	return []byte(body.String()), nil
}
