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
	if len(fileSplit) == 1 {
		logger.Info("No nonce field found. Use NONCEHERE in your file to use it", "nonceCount", len(fileSplit))
	} else if len(fileSplit) == 2 {
		// All Good
		logger.Info("Found one nonce field", "nonceCount", len(fileSplit))
	} else {
		// You probably do not need more than one nonce field
		// If you need it, you can remove this line and/or open a PR
		logger.Warn("MORE THAN ONE NONCE FIELD FOUND", "nonceCount", len(fileSplit))
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
