package canvasx

import (
	"io"
	"mime"
	"mime/multipart"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildImageMultipart(t *testing.T) {
	meta := map[string]any{"title": "x", "size": map[string]any{"width": 1, "height": 2}}
	body, contentType, err := BuildImageMultipart(meta, "foo.png", []byte("PNG-BYTES"))
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(contentType, "multipart/form-data; boundary="))

	_, params, err := mime.ParseMediaType(contentType)
	require.NoError(t, err)
	mr := multipart.NewReader(body, params["boundary"])

	gotParts := map[string]string{}
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		data, err := io.ReadAll(part)
		require.NoError(t, err)
		gotParts[part.FormName()] = string(data)
	}
	assert.Contains(t, gotParts, "json")
	assert.Contains(t, gotParts, "data")
	assert.Contains(t, gotParts["json"], "\"title\":\"x\"")
	assert.Equal(t, "PNG-BYTES", gotParts["data"])
}
