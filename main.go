/**
 * @Author: cyj19
 * @Date: 2022/2/22 20:43
 */
package main

import (
	"log"
	"os"
	"runtime"
)

func main() {

	concurrencyN := runtime.NumCPU()

	app := &cli.App{
		Name:  "downloader",
		Usage: "concurrency download file",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "url",
				Aliases:  []string{"u"},
				Usage:    "`URL` to download",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "`output` to the filename",
			},
			&cli.IntFlag{
				Name:    "concurrency",
				Aliases: []string{"n"},
				Value:   concurrencyN,
				Usage:   "`concurrency` number",
			},
		},
		Action: func(c *cli.Context) error {
			strUrl := c.String("url")
			filename := c.String("output")
			concurrency := c.Int("concurrency")
			return NewDownloader(concurrency).Download(strUrl, filename)
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
