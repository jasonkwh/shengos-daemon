package clipboard

import (
	"bytes"
	"context"
	"crypto/sha256"
	"net/url"
	"strings"

	"golang.design/x/clipboard"
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

func (c *clboard) writeRemembered(
	ctx context.Context,
	format clipboard.Format,
	data []byte,
) error {
	if _, err := clipboard.Write(ctx, format, data); err != nil {
		return err
	}

	c.remember(format, data)

	return nil
}

func (c *clboard) remember(format clipboard.Format, data []byte) {
	hash := clipboardHash(format, data)

	c.mu.Lock()
	c.written = &hash
	c.mu.Unlock()
}

func (c *clboard) ownWrite(format clipboard.Format, data []byte) bool {
	hash := clipboardHash(format, data)

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.written == nil || *c.written != hash {
		return false
	}

	c.written = nil

	return true
}

func clipboardHash(format clipboard.Format, data []byte) [32]byte {
	h := sha256.New()
	h.Write([]byte{byte(format)})
	h.Write(data)

	var sum [32]byte

	copy(sum[:], h.Sum(nil))

	return sum
}
