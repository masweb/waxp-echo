package handler

import (
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"

	_ "image/gif"
	_ "image/png"

	"github.com/chai2010/webp"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/image/draw"
)

const thumbnailSize = 150

func generateThumbnail(srcPath string, mimeType string) (string, error) {
	if mimeType == "image/svg+xml" {
		return "", nil
	}

	f, err := os.Open(srcPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return "", err
	}

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	size := min(w, h)
	cropX := (w - size) / 2
	cropY := (h - size) / 2

	cropped := image.NewRGBA(image.Rect(0, 0, size, size))
	draw.Draw(cropped, cropped.Bounds(), img, image.Pt(cropX, cropY), draw.Src)

	thumb := image.NewRGBA(image.Rect(0, 0, thumbnailSize, thumbnailSize))
	draw.CatmullRom.Scale(thumb, thumb.Bounds(), cropped, cropped.Bounds(), draw.Over, nil)

	base := strings.TrimSuffix(srcPath, filepath.Ext(srcPath))

	dstPath := base + "_thumb.webp"
	out, err := os.Create(dstPath)
	if err != nil {
		return "", err
	}

	err = webp.Encode(out, thumb, &webp.Options{Quality: 80})
	out.Close()
	if err != nil {
		os.Remove(dstPath)

		dstPath = base + "_thumb.jpg"
		out, err = os.Create(dstPath)
		if err != nil {
			return "", err
		}
		defer out.Close()

		err = jpeg.Encode(out, thumb, &jpeg.Options{Quality: 85})
		if err != nil {
			out.Close()
			os.Remove(dstPath)
			return "", err
		}
	}

	return filepath.Base(dstPath), nil
}

func thumbnailURL(mediaDir string, thumbFile string) pgtype.Text {
	if thumbFile == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: "/media/" + thumbFile, Valid: true}
}
