package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const defaultWebAppURL = "https://script.google.com/macros/s/AKfycbyHG3KJJFFRf915NegDb3IGKPCWaEqSDCdJCNBweRcLS5T961uIeNQKD_9tJW1E1sI/exec"

func main() {
	timeout := flag.Duration("timeout", 15*time.Second, "Timeout HTTP")
	flag.Parse()

	webappURL := defaultWebAppURL
	value := "W"

	if err := postToWebApp(webappURL, value, *timeout); err != nil {
		failf("Echec WebApp: %v", err)
	}
	fmt.Println("Succès: valeur envoyée via WebApp.")
}

func postToWebApp(webappURL, value string, timeout time.Duration) error {
	reqBody := strings.NewReader(url.Values{"value": {value}}.Encode())
	req, err := http.NewRequest(http.MethodPost, webappURL, reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("statut HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}
	return nil
}

func failf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", a...)
	os.Exit(1)
}
