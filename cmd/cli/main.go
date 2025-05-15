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
	"strings"
)

const (
	BASE_URL = "http://localhost:8080"
)

func uploadFile(data io.Reader) (string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "data.bin")
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = io.Copy(part, data)
	if err != nil {
		return "", fmt.Errorf("failed to copy file content: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	req, err := http.NewRequest("POST", BASE_URL+"/api/upload", body)
	if err != nil {
		return "", fmt.Errorf("failed to prepare request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Upload-Complete", "1")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("api error (%s) & failed to read response: %w", resp.Status, err)
		}

		return "", fmt.Errorf("api error: %s", string(respBody))
	}

	return resp.Request.URL.String(), nil
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
