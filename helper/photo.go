package helper

import (
	"fmt"
	"sort"
	"strings"
)

func EmbedPhoto(photo string) string {
	photoSplited := strings.Split(photo, "/")
	dataPosition := sort.StringSlice(photoSplited).Search("d")
	dataPosition = dataPosition + 3

	return fmt.Sprint("https://drive.google.com/uc?export=view&id=", photoSplited[dataPosition])
}
