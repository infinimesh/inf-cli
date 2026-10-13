package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func mediaBase(cmd *cobra.Command) string {
	if cmd.Flags().Changed("url") {
		u, _ := cmd.Flags().GetString("url")
		return strings.TrimRight(u, "/")
	}
	if u := viper.GetString("media"); u != "" {
		return strings.TrimRight(u, "/")
	}
	return "https://media.localhost"
}

func mediaRequest(method, rawURL string, body io.Reader, contentType string) (*http.Response, error) {
	client, err := httpClient()
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, rawURL, body)
	if err != nil {
		return nil, err
	}
	if tok := viper.GetString("token"); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return client.Do(req)
}

var mediaCmd = &cobra.Command{
	Use:   "media",
	Short: "Namespace files on http-fs",
	Long: `http-fs is a separate process. Compose serves it at https://media.localhost.

List, upload, and delete send the same account token as the API. A downloaded file is public.
Pass --url or store media in the CLI config. TLS uses --ca-file and --insecure-skip-verify.`,
}

var listMediaCmd = &cobra.Command{
	Use:   "list <namespace>",
	Short: "List files",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 || args[0] == "" {
			return errors.New("namespace is required")
		}
		res, err := mediaRequest(http.MethodGet, mediaBase(cmd)+"/"+args[0], nil, "")
		if err != nil {
			return err
		}
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}
		if res.StatusCode >= 300 {
			return fmt.Errorf("%s: %s", res.Status, bytes.TrimSpace(body))
		}
		if printJson, _ := cmd.Flags().GetBool("json"); printJson {
			fmt.Println(string(body))
			return nil
		}
		var page struct {
			Files []struct {
				Name    string `json:"name"`
				Size    int64  `json:"size"`
				ModTime string `json:"mod_time"`
			} `json:"files"`
		}
		if err := json.Unmarshal(body, &page); err != nil {
			fmt.Println(string(body))
			return nil
		}
		for _, f := range page.Files {
			fmt.Printf("%s\t%d\t%s\n", f.Name, f.Size, f.ModTime)
		}
		return nil
	},
}

var putMediaCmd = &cobra.Command{
	Use:   "put <namespace> <file>",
	Short: "Upload a file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("namespace and file are required")
		}
		f, err := os.Open(args[1])
		if err != nil {
			return err
		}
		defer f.Close()
		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		part, err := w.CreateFormFile("file", path.Base(args[1]))
		if err != nil {
			return err
		}
		if _, err = io.Copy(part, f); err != nil {
			return err
		}
		if err = w.Close(); err != nil {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			name = path.Base(args[1])
		}
		res, err := mediaRequest(http.MethodPost, mediaBase(cmd)+"/"+args[0]+"/"+name, &buf, w.FormDataContentType())
		if err != nil {
			return err
		}
		defer res.Body.Close()
		body, _ := io.ReadAll(res.Body)
		if res.StatusCode >= 300 {
			return fmt.Errorf("%s: %s", res.Status, bytes.TrimSpace(body))
		}
		fmt.Println(name)
		return nil
	},
}

var rmMediaCmd = &cobra.Command{
	Use:   "rm <namespace> <name>",
	Short: "Delete a file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("namespace and name are required")
		}
		res, err := mediaRequest(http.MethodDelete, mediaBase(cmd)+"/"+args[0]+"/"+args[1], nil, "")
		if err != nil {
			return err
		}
		defer res.Body.Close()
		body, _ := io.ReadAll(res.Body)
		if res.StatusCode >= 300 {
			return fmt.Errorf("%s: %s", res.Status, bytes.TrimSpace(body))
		}
		fmt.Println("deleted", args[1])
		return nil
	},
}

var getMediaCmd = &cobra.Command{
	Use:   "get <namespace> <name>",
	Short: "Download a public file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("namespace and name are required")
		}
		res, err := mediaRequest(http.MethodGet, mediaBase(cmd)+"/"+args[0]+"/"+args[1], nil, "")
		if err != nil {
			return err
		}
		defer res.Body.Close()
		if res.StatusCode >= 300 {
			body, _ := io.ReadAll(res.Body)
			return fmt.Errorf("%s: %s", res.Status, bytes.TrimSpace(body))
		}
		out, _ := cmd.Flags().GetString("output")
		if out == "" || out == "-" {
			_, err = io.Copy(os.Stdout, res.Body)
			return err
		}
		f, err := os.Create(out)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(f, res.Body)
		return err
	},
}

func init() {
	for _, c := range []*cobra.Command{listMediaCmd, putMediaCmd, rmMediaCmd, getMediaCmd} {
		c.Flags().String("url", "", "http-fs base URL (default https://media.localhost)")
	}
	putMediaCmd.Flags().String("name", "", "Stored name. Defaults to the file name")
	getMediaCmd.Flags().StringP("output", "o", "", "Write the file here. Empty prints it")
	mediaCmd.AddCommand(listMediaCmd, putMediaCmd, rmMediaCmd, getMediaCmd)
	rootCmd.AddCommand(mediaCmd)
}
