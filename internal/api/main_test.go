package api

import (
	"io"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/stretchr/testify/require"
	"github.com/yosssi/gohtml"
)

// snapshotHTML snapshots the given HTML with formatting.
func snapshotHTML(t *testing.T, body io.ReadCloser) {
	t.Helper()

	b, err := io.ReadAll(body)
	require.NoError(t, err)
	defer body.Close()

	cupaloy.New(cupaloy.SnapshotFileExtension(".html")).
		SnapshotT(t, string(gohtml.FormatBytes(b)))
}
