package server

import (
	"log"
	"os"
	"strings"

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
		log.Println("[INFO] No nonce field found. Use NONCEHERE in your file to use it")
	} else if len(fileSplit) == 2 {
		// All Good
		log.Println("[INFO] Found one nonce field")
	} else {
		// You probably do not need more than one nonce field
		// If you need it, you can remove this line and/or open a PR
		log.Println("[WARN] MORE THAN ONE NONCE FIELD FOUND")
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
