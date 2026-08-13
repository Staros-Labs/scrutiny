package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var bundledBinaries = []string{
	"scrutiny-collector-omnibus",
	"scrutiny-collector-metrics",
	"scrutiny-collector-zfs",
	"scrutiny-collector-mdadm",
	"scrutiny-collector-btrfs",
	"scrutiny-collector-filesystem",
	"scrutiny-collector-performance",
}

var bundledFiles = []archiveFile{
	{source: "example.collector-omnibus.yaml", destination: "config/collector-omnibus.yaml", mode: 0o644},
	{source: "example.collector.yaml", destination: "config/collector.yaml", mode: 0o644},
	{source: "example.collector-zfs.yaml", destination: "config/collector-zfs.yaml", mode: 0o644},
	{source: "example.collector-btrfs.yaml", destination: "config/collector-btrfs.yaml", mode: 0o644},
	{source: "example.collector-performance.yaml", destination: "config/collector-performance.yaml", mode: 0o644},
	{source: "docs/INSTALL_COLLECTOR_OMNIBUS.md", destination: "INSTALL.md", mode: 0o644},
	{source: "LICENSE", destination: "LICENSE", mode: 0o644},
}

var archiveTimestamp = time.Date(1980, time.January, 2, 0, 0, 0, 0, time.UTC)

type archiveFile struct {
	source      string
	destination string
	mode        os.FileMode
}

type packageOptions struct {
	sourceDir  string
	outputDir  string
	targetOS   string
	targetArch string
	targetARM  string
}

func main() {
	var options packageOptions
	flag.StringVar(&options.sourceDir, "source-dir", ".", "repository root containing built binaries")
	flag.StringVar(&options.outputDir, "output-dir", ".", "directory for the generated archive")
	flag.StringVar(&options.targetOS, "target-os", "", "target GOOS")
	flag.StringVar(&options.targetArch, "target-arch", "", "target GOARCH")
	flag.StringVar(&options.targetARM, "target-arm", "", "target GOARM")
	flag.Parse()

	if options.targetOS == "" || options.targetArch == "" {
		fmt.Fprintln(os.Stderr, "target-os and target-arch are required")
		os.Exit(2)
	}
	archivePath, err := createArchive(&options)
	if err != nil {
		fmt.Fprintf(os.Stderr, "package collector omnibus: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(archivePath)
}

func createArchive(options *packageOptions) (string, error) {
	platform := options.targetOS + "-" + options.targetArch
	if options.targetARM != "" {
		platform += "-" + options.targetARM
	}
	root := "scrutiny-collector-omnibus-" + platform
	files := make([]archiveFile, 0, len(bundledBinaries)+len(bundledFiles))
	for _, name := range bundledBinaries {
		sourceName := name + "-" + platform
		destinationName := name
		if options.targetOS == "windows" {
			sourceName += ".exe"
			destinationName += ".exe"
		}
		files = append(files, archiveFile{
			source:      sourceName,
			destination: "bin/" + destinationName,
			mode:        0o755,
		})
	}
	files = append(files, bundledFiles...)

	if err := os.MkdirAll(options.outputDir, 0o755); err != nil {
		return "", fmt.Errorf("create output directory: %w", err)
	}
	if options.targetOS == "windows" {
		path := filepath.Join(options.outputDir, root+".zip")
		return path, writeZip(path, options.sourceDir, root, files)
	}
	path := filepath.Join(options.outputDir, root+".tar.gz")
	return path, writeTarGz(path, options.sourceDir, root, files)
}

func writeTarGz(outputPath, sourceDir, root string, files []archiveFile) (returnErr error) {
	output, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer closeWithError(output, &returnErr)

	gzipWriter := gzip.NewWriter(output)
	defer closeWithError(gzipWriter, &returnErr)
	tarWriter := tar.NewWriter(gzipWriter)
	defer closeWithError(tarWriter, &returnErr)

	for _, file := range files {
		data, err := os.ReadFile(filepath.Join(sourceDir, file.source))
		if err != nil {
			return fmt.Errorf("read %s: %w", file.source, err)
		}
		header := &tar.Header{
			Name:    filepath.ToSlash(filepath.Join(root, file.destination)),
			Mode:    int64(file.mode.Perm()),
			Size:    int64(len(data)),
			ModTime: archiveTimestamp,
		}
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}
		if _, err := tarWriter.Write(data); err != nil {
			return err
		}
	}
	return nil
}

func writeZip(outputPath, sourceDir, root string, files []archiveFile) (returnErr error) {
	output, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer closeWithError(output, &returnErr)

	zipWriter := zip.NewWriter(output)
	defer closeWithError(zipWriter, &returnErr)

	for _, file := range files {
		data, err := os.ReadFile(filepath.Join(sourceDir, file.source))
		if err != nil {
			return fmt.Errorf("read %s: %w", file.source, err)
		}
		header := &zip.FileHeader{
			Name:   strings.ReplaceAll(filepath.Join(root, file.destination), string(filepath.Separator), "/"),
			Method: zip.Deflate,
		}
		header.SetMode(file.mode)
		header.Modified = archiveTimestamp
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}
		if _, err := writer.Write(data); err != nil {
			return err
		}
	}
	return nil
}

func closeWithError(closer io.Closer, returnErr *error) {
	if err := closer.Close(); err != nil && !errors.Is(err, os.ErrClosed) && *returnErr == nil {
		*returnErr = err
	}
}
