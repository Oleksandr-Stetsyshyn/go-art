package photoprocessor

import (
	"fmt"
	"github.com/disintegration/imaging"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

type PhotoProcessor interface {
	SavePhotos(files []*multipart.FileHeader, path string) error
	ResizePhotos(inputPath string, outputPath string, newLongSide int) error
	RemoveFolder(path string) error
}

type LocalPhotoProcessor struct{}

func (p *LocalPhotoProcessor) SavePhotos(files []*multipart.FileHeader, path string) error {

	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, 0755)
		if err != nil {
			return err
		}
	}

	for _, file := range files {
		src, err := file.Open()
		if err != nil {
			return err
		}
		defer src.Close()

		dst, err := os.CreateTemp(path, "*.jpg")
		if err != nil {
			return err
		}
		defer dst.Close()

		_, err = io.Copy(dst, src)
		if err != nil {
			return err
		}
	}

	return nil
}
func (p *LocalPhotoProcessor) ResizePhotos(inputPath string, outputPath string, newLongSide int) error {
	outputDir := filepath.Dir(outputPath)
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		err = os.MkdirAll(outputDir, 0755)
		if err != nil {
			return err
		}
	}

	img, err := imaging.Open(inputPath)
	if err != nil {
		return err
	}

	var newWidth, newHeight int
	if img.Bounds().Dx() > img.Bounds().Dy() {
		newWidth = newLongSide
		newHeight = 0
	} else {
		newWidth = 0
		newHeight = newLongSide
	}

	resizedImg := imaging.Resize(img, newWidth, newHeight, imaging.CatmullRom)

	err = imaging.Save(resizedImg, outputPath)
	if err != nil {
		return err
	}

	return nil
}
func (p *LocalPhotoProcessor) RemoveFolder(tempFolderPath string) error {
	if _, err := os.Stat(tempFolderPath); !os.IsNotExist(err) {
		err = os.RemoveAll(tempFolderPath)
		if err != nil {
			return err
		} else {
			fmt.Println("Temporary folder removed successfully")
		}
	}

	return nil
}

func SaveAndResizeFiles(processor PhotoProcessor, files []*multipart.FileHeader, savePath string, resizePath string) error {

	err := processor.SavePhotos(files, savePath)
	if err != nil {
		return err
	}

	savedFiles, err := os.ReadDir(savePath)
	if err != nil {
		return err
	}

	for _, savedFile := range savedFiles {

		if !savedFile.Type().IsRegular() {
			continue
		}

		inputPath := filepath.Join(savePath, savedFile.Name())
		outputPath := filepath.Join(resizePath, savedFile.Name())

		err = processor.ResizePhotos(inputPath, outputPath, 1500)
		if err != nil {
			return err
		}
	}

	return nil
}
