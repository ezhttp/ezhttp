package server

import (
	"os"
	"strings"

	"github.com/ezhttp/ezhttp/internal/logger"
	"github.com/tdewolff/minify/v2"
)

// CachedIndexString holds the split index.html file
var CachedIndexString []string

// LoadIndexCache loads and splits the index.html file
func LoadIndexCache() ([]string, error) {
	fileBytes, _ := os.ReadFile("./public/index.html")
	fileString := string(fileBytes)
	fileSplit := strings.Split(fileString, "NONCEHERE")
	nonceFieldCount := len(fileSplit) - 1
	if nonceFieldCount == 0 {
		logger.Info("No nonce field found. Use NONCEHERE in your file to use it", "nonceCount", nonceFieldCount)
	} else if nonceFieldCount <= 2 {
		// Expected: 1-2 nonces (style, script)
		logger.Info("Found nonce fields", "nonceCount", nonceFieldCount)
	} else {
		// More than 2 nonces is unusual and might indicate an issue
		logger.Warn("Unusually high number of nonce fields found", "nonceCount", nonceFieldCount, "expected", "1-2 (style, script)")
	}

	//bufio.NewScanner()
	return fileSplit, nil
}

// GenerateIndexWithNonce generates the index HTML with the nonce inserted
func GenerateIndexWithNonce(nonce string, cachedIndexString []string, minifier *minify.M, banner []string) string {
	totalSlots := len(cachedIndexString)
	finalReturn := make([]string, (totalSlots*2)-1)
	for i := 0; i < totalSlots; i++ {
		finalReturn[(i * 2)] = cachedIndexString[i]
		if (i + 1) != len(cachedIndexString) {
			finalReturn[(i*2)+1] = nonce
		}
	}
	finalString := strings.Join(finalReturn, "")
	minified, _ := minifier.String("text/html", finalString)
	minifiedWithBanner := strings.Replace(minified, "</head>", strings.Join(banner, "\n")+"</head>", 1)
	minifiedAngularNonce := strings.Replace(minifiedWithBanner, "ngcspnonce", "ngCspNonce", 1)
	return minifiedAngularNonce
	//buf := &bytes.Buffer{}
	// import "encoding/gob"
	//gob.NewEncoder(buf).Encode(strings.Join(finalReturn, ""))
	//bs := buf.Bytes()
	//return bs
}
