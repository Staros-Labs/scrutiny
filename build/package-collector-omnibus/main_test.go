package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateArchiveUnixIncludesCanonicalBundleLayout(t *testing.T) {
	sourceDir := packageFixture(t, "linux-amd64", "")
	outputDir := t.TempDir()

	path, err := createArchive(&packageOptions{
		sourceDir:  sourceDir,
		outputDir:  outputDir,
		targetOS:   "linux",
		targetArch: "amd64",
	})
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(outputDir, "scrutiny-collector-omnibus-linux-amd64.tar.gz"), path)

	files := readTarFiles(t, path)
	root := "scrutiny-collector-omnibus-linux-amd64/"
	for _, binary := range bundledBinaries {
		assert.Contains(t, files, root+"bin/"+binary)
		assert.Equal(t, int64(0o755), files[root+"bin/"+binary])
	}
	assert.Contains(t, files, root+"config/collector-omnibus.yaml")
	assert.Contains(t, files, root+"INSTALL.md")
	assert.Contains(t, files, root+"LICENSE")
}

func TestCreateArchiveWindowsUsesZipAndExeNames(t *testing.T) {
	sourceDir := packageFixture(t, "windows-arm64", ".exe")
	outputDir := t.TempDir()

	path, err := createArchive(&packageOptions{
		sourceDir:  sourceDir,
		outputDir:  outputDir,
		targetOS:   "windows",
		targetArch: "arm64",
	})
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(outputDir, "scrutiny-collector-omnibus-windows-arm64.zip"), path)

	reader, err := zip.OpenReader(path)
	require.NoError(t, err)
	defer reader.Close()
	names := make(map[string]struct{}, len(reader.File))
	for _, file := range reader.File {
		names[file.Name] = struct{}{}
	}
	root := "scrutiny-collector-omnibus-windows-arm64/"
	for _, binary := range bundledBinaries {
		assert.Contains(t, names, root+"bin/"+binary+".exe")
	}
}

func packageFixture(t *testing.T, platform, extension string) string {
	t.Helper()
	dir := t.TempDir()
	for _, binary := range bundledBinaries {
		require.NoError(t, os.WriteFile(filepath.Join(dir, binary+"-"+platform+extension), []byte(binary), 0o755))
	}
	for _, file := range bundledFiles {
		path := filepath.Join(dir, file.source)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte(file.source), file.mode))
	}
	return dir
}

func readTarFiles(t *testing.T, path string) map[string]int64 {
	t.Helper()
	file, err := os.Open(path)
	require.NoError(t, err)
	defer file.Close()
	gzipReader, err := gzip.NewReader(file)
	require.NoError(t, err)
	defer gzipReader.Close()
	reader := tar.NewReader(gzipReader)
	files := map[string]int64{}
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		files[header.Name] = header.Mode
	}
	return files
}
