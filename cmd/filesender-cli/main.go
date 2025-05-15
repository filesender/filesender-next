// Cli for filesender
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	baseURL   = "http://localhost:8080"
	chunkSize = 1024 * 1024
)

func uploadFile(data io.Reader) (string, error) {
	uploadMethod := "POST"
	uploadDesitionation := baseURL + "/api/upload"
	buf := make([]byte, chunkSize)
	var offset int64

	for {
		n, err := io.ReadFull(data, buf)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return "", fmt.Errorf("failed to read chunk: %w", err)
		}

		isLastChunk := false
		if n < chunkSize {
			isLastChunk = true
		}

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		part, err := writer.CreateFormFile("file", "data.bin")
		if err != nil {
			return "", fmt.Errorf("failed to create form file: %w", err)
		}

		_, err = part.Write(buf[:n])
		if err != nil {
			return "", fmt.Errorf("failed to write chunk to part: %w", err)
		}

		err = writer.Close()
		if err != nil {
			return "", fmt.Errorf("failed to close writer: %w", err)
		}

		req, err := http.NewRequest(uploadMethod, uploadDesitionation, body)
		if err != nil {
			return "", fmt.Errorf("failed to prepare request: %w", err)
		}

		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Upload-Complete", "0")

		if isLastChunk {
			req.Header.Set("Upload-Complete", "1")
		}

		if offset > 0 {
			req.Header.Set("Upload-Offset", strconv.FormatInt(offset, 10))
		}

		tries := 0
		for {
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				if tries >= 3 {
					return "", fmt.Errorf("failed to make request: %w", err)
				}

				tries++
				time.Sleep(3 * time.Second)
				continue
			}
			defer func() {
				err := resp.Body.Close()
				if err != nil {
					fmt.Printf("Error: %s\n", err)
				}
			}()

			if resp.StatusCode == 200 {
				return resp.Request.URL.String(), nil
			}

			if resp.StatusCode == 202 {
				uploadMethod = "PATCH"
				uploadDesitionation = baseURL + resp.Header.Get("Location")
				offset += int64(n)
				break
			}

			if tries >= 3 {
				respBody, err := io.ReadAll(resp.Body)
				if err != nil {
					return "", fmt.Errorf("api error after 3 tries (%s) & failed to read response: %w", resp.Status, err)
				}

				return "", fmt.Errorf("api error after 3 tries (%d): %s", resp.StatusCode, string(respBody))
			}

			tries++
			time.Sleep(3 * time.Second)
		}
	}
}

func downloadFile(link string) (io.ReadCloser, error) {
	pr, pw := io.Pipe()

	go func() {
		var offset int64
		tries := 0

		for {
			req, err := http.NewRequest("GET", link, nil)
			if err != nil {
				pw.CloseWithError(err)
				return
			}
			if offset > 0 {
				req.Header.Set("Range", fmt.Sprintf("bytes=%d-", offset))
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				if tries >= 3 {
					pw.CloseWithError(fmt.Errorf("failed downloading after three tries: %w", err))
					return
				}

				time.Sleep(3 * time.Second)
				tries++
				continue
			}
			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
				err := resp.Body.Close()
				if err != nil {
					fmt.Printf("Error: %s\n", err)
				}

				pw.CloseWithError(fmt.Errorf("unexpected status: %s", resp.Status))
				return
			}
			tries = 0

			buf := make([]byte, 32*1024)
			for {
				n, err := resp.Body.Read(buf)
				if n > 0 {
					wn, werr := pw.Write(buf[:n])
					if werr != nil {
						err = resp.Body.Close()
						if err != nil {
							fmt.Printf("Error: %s\n", err)
						}

						pw.CloseWithError(werr)
						return
					}
					offset += int64(wn)
				}
				if err != nil {
					cerr := resp.Body.Close()
					if cerr != nil {
						fmt.Printf("Error: %s\n", cerr)
					}

					if err == io.EOF {
						err = pw.Close()
						if err != nil {
							fmt.Printf("Error: %s\n", err)
						}
						return
					}
					break
				}
			}
		}
	}()

	return pr, nil
}

func upload(_ bool, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		}
	}()

	uploadLocation, err := uploadFile(file)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	fmt.Printf("Uploaded here: %s\n", uploadLocation)
	return nil
}

func download(link string, filePath string) error {
	link = strings.Replace(link, "/view/", "/download/", 1)

	// parts := strings.Split(link, "#")
	// encrypted := false
	// keys := ""
	// if len(parts) == 2 {
	// 	encrypted = true
	// 	link, keys = parts[0], parts[1]
	// }

	reader, err := downloadFile(link)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer func() {
		err := reader.Close()
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		}
	}()

	if filePath == "" {
		_, err := io.Copy(os.Stdout, reader)
		if err != nil {
			return fmt.Errorf("failed showing file: %w", err)
		}
		return nil
	}

	return nil
}

func main() {
	if len(os.Args) < 2 {
		// TODO: usage
		log.Fatalf("expected 'upload' or 'download' subcommands")
	}

	uploadCmd := flag.NewFlagSet("upload", flag.ExitOnError)
	uploadSecure := uploadCmd.Bool("s", false, "If secure (encrypted) upload")

	downloadCmd := flag.NewFlagSet("download", flag.ExitOnError)
	downloadOutputFile := downloadCmd.String("o", "", "Output file")

	switch os.Args[1] {
	case "upload":
		err := uploadCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			os.Exit(1)
		}

		if len(uploadCmd.Args()) == 0 {
			log.Fatalf("No file specified for upload")
		}

		err = upload(*uploadSecure, uploadCmd.Arg(0))
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			os.Exit(1)
		}
	case "download":
		err := downloadCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			os.Exit(1)
		}

		if len(downloadCmd.Args()) == 0 {
			log.Fatalf("Download command requires a URL")
		}

		err = download(downloadCmd.Arg(0), *downloadOutputFile)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			os.Exit(1)
		}
	default:
		log.Fatalf("Unknown command: %s", os.Args[0])
	}

}
