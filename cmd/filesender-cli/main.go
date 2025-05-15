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
)

const (
	BASE_URL   = "http://localhost:8080"
	CHUNK_SIZE = 1024 * 1024
)

func uploadFile(data io.Reader) (string, error) {
	uploadMethod := "POST"
	uploadDesitionation := BASE_URL + "/api/upload"
	buf := make([]byte, CHUNK_SIZE)
	offset := 0

	for {
		n, err := io.ReadFull(data, buf)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return "", fmt.Errorf("failed to read chunk: %w", err)
		}

		isLastChunk := false
		if n < CHUNK_SIZE {
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
			req.Header.Set("Upload-Offset", strconv.FormatInt(int64(offset), 10))
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return "", fmt.Errorf("failed to make request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			return resp.Request.URL.String(), nil
		}

		if resp.StatusCode == 202 {
			uploadMethod = "PATCH"
			uploadDesitionation = BASE_URL + resp.Header.Get("Location")
			offset += n
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("api error (%s) & failed to read response: %w", resp.Status, err)
		}

		return "", fmt.Errorf("api error (%d): %s", resp.StatusCode, string(respBody))
	}
}

func downloadFile(link string) (io.ReadCloser, error) {
	resp, err := http.Get(link)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("api error (%s) & failed to read response: %w", resp.Status, err)
		}

		resp.Body.Close()
		return nil, fmt.Errorf("bad status (%d): %s", resp.StatusCode, respBody)
	}

	return resp.Body, nil
}

func upload(secure bool, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

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
	defer reader.Close()

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
		uploadCmd.Parse(os.Args[2:])
		if len(uploadCmd.Args()) == 0 {
			log.Fatalf("No file specified for upload")
		}
		if err := upload(*uploadSecure, uploadCmd.Arg(0)); err != nil {
			fmt.Printf("Error: %s\n", err)
			os.Exit(1)
		}
	case "download":
		downloadCmd.Parse(os.Args[2:])
		if len(downloadCmd.Args()) == 0 {
			log.Fatalf("Download command requires a URL")
		}
		if err := download(downloadCmd.Arg(0), *downloadOutputFile); err != nil {
			fmt.Printf("Error: %s\n", err)
			os.Exit(1)
		}
	default:
		log.Fatalf("Unknown command: %s", os.Args[0])
	}

}
