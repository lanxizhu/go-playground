package utils

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/disintegration/imaging"
)

// FitSize Reduced ratio of the original image
const FitSize = 16

func getExtension(filename string) (name string, extension string) {
	lastDot := strings.LastIndex(filename, ".")
	if lastDot == -1 || lastDot == 0 {
		return filename, ""
	}
	return filename[:lastDot], filename[lastDot+1:]

	// pattern := `\.([^.]+)$`

	// r := regexp.MustCompile(pattern)
	// matches := r.FindStringSubmatch(filename)
	// if len(matches) > 1 {
	// 	extension = matches[1] // Returns extension without dot
	// 	return
	// }
	// return ""
}

// FitImage Resizes the image to a smaller size
func FitImage(filePath string, fitSize int) (err error) {
	pattern := `\.(jpg|jpeg|png|gif)$`

	if fitSize == 0 {
		fitSize = FitSize
	}

	r := regexp.MustCompile(pattern)

	if !r.MatchString(filePath) {
		return fmt.Errorf("file is not an image")
	}

	content, err := os.Open(filePath)
	if err != nil {
		return err
	}

	name, extension := getExtension(content.Name())

	defer func(content *os.File) {
		err = content.Close()
		if err != nil {
			return
		}
	}(content)

	buffer := make([]byte, 1024)
	_, err = content.Read(buffer)
	if err != nil {
		return err
	}

	fileType := http.DetectContentType(buffer)

	fmt.Println("File type:", fileType)

	dst, err := imaging.Open(filePath)

	if err != nil {
		return err
	}

	size := dst.Bounds().Size()

	// dst = imaging.Resize(dst, size.X/FitSize, size.Y/FitSize, imaging.Lanczos)
	dstImageFit := imaging.Fit(dst, size.X/fitSize, size.Y/fitSize, imaging.Box)
	fmt.Printf("Original Size: %s, Fit size: %s\n", dst.Bounds().Size(), dstImageFit.Bounds().Size())
	err = imaging.Save(dstImageFit, fmt.Sprintf("%s_fit.%s", name, extension))

	return err
}
