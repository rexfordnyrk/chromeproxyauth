package proxyauth

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
)

func BuildExtention(host, port, userName, password string) (string, error) {

	extZipFile := "ProxyAuth.zip"

	const (
		manifestFName = "manifest.json"
		backFName     = "background.js"
	)

	manifest_json := `{
	  "version": "1.0.0",
	  "manifest_version": 2,
	  "name": "Proxy Auth",
	  "permissions": [
	    "proxy",
	    "tabs",
	    "unlimitedStorage",
	    "storage",
	    "<all_urls>",
	    "webRequest",
	    "webRequestBlocking"
	  ],
	  "background": {
	    "scripts": ["background.js"]
	  },
	  "minimum_chrome_version":"22.0.0"
	}`

	background_js := fmt.Sprintf(`var config = {
	  mode: "fixed_servers",
	  rules: {
	    singleProxy: {
	      scheme: "http",
	      host: "%s",
	      port: parseInt(%s)
	    },
	    bypassList: ["localhost"]
	  }
	};
	
	chrome.proxy.settings.set({value: config, scope: "regular"}, function() {});
	
	function callbackFn(details) {
	  return {
	    authCredentials: {
	      username: "%s",
	      password: "%s"
	    }
	  };
	}
	
	chrome.webRequest.onAuthRequired.addListener(
	  callbackFn,
	  {urls: ["<all_urls>"]},
	  ['blocking']
	);`, host, port, userName, password)

	fos, err := os.Create(extZipFile)
	if err != nil {
		return "", fmt.Errorf("os.Create Error: %w", err)
	}
	defer fos.Close()
	zipWriter := zip.NewWriter(fos)

	preManifestFile, err := os.Create(manifestFName)
	if err != nil {
		return "", fmt.Errorf("os.Create manifestFile Error: %w", err)
	}
	if _, err = preManifestFile.Write([]byte(manifest_json)); err != nil {
		return "", fmt.Errorf("preManifestFile.Write Error: %w", err)
	}
	preManifestFile.Close()
	manifestFile, err := os.Open(manifestFName)
	if err != nil {
		return "", fmt.Errorf("os.Open manifestFile  Error: %w", err)
	}

	wmf, err := zipWriter.Create(manifestFName)
	if err != nil {
		return "", fmt.Errorf("zipWriter.Create Error: %w", err)
	}
	if _, err := io.Copy(wmf, manifestFile); err != nil {
		return "", fmt.Errorf("io.Copy Error: %w", err)
	}
	manifestFile.Close()

	preBackFile, err := os.Create(backFName)
	if err != nil {
		return "", fmt.Errorf("os.Create preBackFile  Error: %w", err)
	}
	if _, err = preBackFile.Write([]byte(background_js)); err != nil {
		return "", fmt.Errorf("preBackFile.Write Error: %w", err)
	}
	preBackFile.Close()

	backFile, err := os.Open(backFName)
	if err != nil {
		return "", fmt.Errorf("os.Open backFile Error: %w", err)
	}

	wbf, err := zipWriter.Create(backFName)
	if err != nil {
		return "", fmt.Errorf("zipWriter.Create(backFName) Error: %w", err)
	}

	if _, err := io.Copy(wbf, backFile); err != nil {
		return "", fmt.Errorf("io.Copy Error: %w", err)
	}
	backFile.Close()

	if err := os.Remove(manifestFName); err != nil {
		return "", fmt.Errorf("os.Remove(manifestFName) Error: %w", err)
	}

	if err := os.Remove(backFName); err != nil {
		return "", fmt.Errorf("os.Remove(backFName) Error: %w", err)
	}

	return extZipFile, zipWriter.Close()
}
