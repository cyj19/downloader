/**
 * @Author: cyj19
 * @Date: 2022/2/22 20:46
 */
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
)

type Downloader struct {
	concurrency int
}

func NewDownloader(concurrency int) *Downloader {
	return &Downloader{concurrency: concurrency}
}

func (d *Downloader) Download(strUrl, filename string) error {
	if filename == "" {
		filename = path.Base(strUrl)
	}

	resp, err := http.Head(strUrl)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK && resp.Header.Get("Accept-Ranges") == "bytes" {
		return d.multiDownload(strUrl, filename, int(resp.ContentLength))
	}

	return d.singleDownload(strUrl, filename)
}

func (d *Downloader) multiDownload(strUrl, filename string, contentLen int) error {
	partSize := contentLen / d.concurrency

	partDir := d.getPartDir(filename)
	os.Mkdir(partDir, 0777)
	defer os.RemoveAll(partDir)

	var wg sync.WaitGroup
	wg.Add(d.concurrency)
	rangeStart := 0
	log.Println("正在下载>>>")
	for i := 0; i < d.concurrency; i++ {
		// 并发下载
		go func(rangeStart, i int) {
			defer wg.Done()
			rangeEnd := rangeStart + partSize
			if i == d.concurrency-1 {
				rangeEnd = contentLen
			}
			d.downloadPartial(strUrl, filename, rangeStart, rangeEnd, i)
		}(rangeStart, i)

		rangeStart += partSize + 1
	}

	wg.Wait()

	// 合并文件
	return d.merge(filename)
}

func (d *Downloader) singleDownload(strUrl, filename string) error {
	return nil
}

func (d *Downloader) getPartDir(filename string) string {
	return strings.SplitN(filename, ".", 2)[0]
}

func (d *Downloader) getPartFilename(filename string, i int) string {
	pathDir := d.getPartDir(filename)
	return fmt.Sprintf("%s/%s-%d", pathDir, filename, i)
}

func (d *Downloader) downloadPartial(strUrl, filename string, rangeStart, rangeEnd, i int) {
	if rangeStart >= rangeEnd {
		return
	}

	partFile, err := os.OpenFile(d.getPartFilename(filename, i), os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer partFile.Close()

	req, err := http.NewRequest("GET", strUrl, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", rangeStart, rangeEnd))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	buf := make([]byte, 32*1024)

	_, err = io.CopyBuffer(partFile, resp.Body, buf)
	if err != nil {
		if err == io.EOF {
			return
		}
		log.Fatal(err)
	}
}

func (d *Downloader) merge(filename string) error {
	destFile, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		return err
	}
	defer destFile.Close()

	for i := 0; i < d.concurrency; i++ {
		partFilename := d.getPartFilename(filename, i)
		partFile, err := os.Open(partFilename)
		if err != nil {
			return err
		}
		io.Copy(destFile, partFile)
		partFile.Close()
		os.Remove(partFilename)
	}

	return nil
}
