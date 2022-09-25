package cmd

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/dustin/go-humanize"
	"github.com/mitchellh/ioprogress"
)

func GenericPut(debug bool, src, dst string) (err error) {

	var chunkSize int64 = 1 << 24
	workers := 4

	contents, err := os.Open(src)
	if err != nil {
		return
	}
	defer contents.Close()

	contentsInfo, err := contents.Stat()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	progressbar := &ioprogress.Reader{
		Reader: contents,
		DrawFunc: ioprogress.DrawTerminalf(os.Stderr, func(progress, total int64) string {
			return fmt.Sprintf("Uploading %s/%s",
				humanize.IBytes(uint64(progress)), humanize.IBytes(uint64(total)))
		}),
		Size: contentsInfo.Size(),
	}

	commitInfo := files.NewCommitInfo(dst)
	commitInfo.Mode.Tag = "overwrite"

	// The Dropbox API only accepts timestamps in UTC with second precision.
	ts := time.Now().UTC().Round(time.Second)
	commitInfo.ClientModified = &ts

	dbx := files.New(config)
	if contentsInfo.Size() > singleShotUploadSizeCutoff {
		return uploadChunked(dbx, progressbar, commitInfo, contentsInfo.Size(), workers, chunkSize, debug)
	}

	if _, err = dbx.Upload(commitInfo, progressbar); err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(fmt.Sprintf("Upload of %s completed", dst))

	e := os.Remove(src)
	if e != nil {
		fmt.Println(fmt.Sprintf("Can't delete %s", src))
	} else {
		fmt.Println(fmt.Sprintf("File %s deleted from the disk", src))
	}
	return
}

func Tar(source, target string) error {
	tarfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer tarfile.Close()

	tarball := tar.NewWriter(tarfile)
	defer tarball.Close()

	info, err := os.Stat(source)
	if err != nil {
		return nil
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	return filepath.Walk(source,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			header, err := tar.FileInfoHeader(info, info.Name())
			if err != nil {
				return err
			}

			if baseDir != "" {
				header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
			}

			if err := tarball.WriteHeader(header); err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(tarball, file)
			return err
		})
}
