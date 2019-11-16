package processor

import "gopkg.in/gographics/imagick.v2/imagick"

func ThumbnailProcessor(body []byte) error {
	imagick.Initialize()
	defer imagick.Terminate()

	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	return nil
}
