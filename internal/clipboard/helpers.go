package clipboard

import (
	"bytes"
	"context"
	"crypto/sha256"
	"net/url"
	"strings"

	"golang.design/x/clipboard"
)

func (c *clboard) toSFTPURIList(
	origin string,
	data []byte,
) ([]byte, error) {
	var body strings.Builder

	for line := range bytes.SplitSeq(data, []byte("\n")) {
		line = bytes.TrimRight(line, "\r")

		if len(line) == 0 || line[0] == '#' {
			continue
		}

		fileURL, err := url.Parse(string(line))
		if err != nil || fileURL.Scheme != "file" || fileURL.Host != "" || len(fileURL.Path) == 0 || fileURL.Path[0] != '/' {
			continue
		}

		sftpURL := url.URL{
			Scheme: "sftp",
			User:   url.User(c.cfg.SSHUser),
			Host:   origin,
			Path:   fileURL.Path,
		}

		body.WriteString(sftpURL.String())
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
	hash := sha256.New()
	hash.Write([]byte{byte(toMessageClipboardFormat(format))})
	hash.Write(data)

	var sum [32]byte

	copy(sum[:], hash.Sum(nil))

	return sum
}
