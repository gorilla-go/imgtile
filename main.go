package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	inputFolder string
	columns     int
	outputFile  string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "imgtile",
		Short: "Merge images from a folder into one image",
		Run: func(cmd *cobra.Command, args []string) {
			if inputFolder == "" || columns <= 0 || outputFile == "" {
				log.Fatal("Please provide -i (input folder), -c (columns), and -o (output file)")
			}
			mergeImages(inputFolder, columns, outputFile)
		},
	}

	rootCmd.Flags().StringVarP(&inputFolder, "input", "i", "", "Input folder with images")
	rootCmd.Flags().IntVarP(&columns, "columns", "c", 6, "Number of images per row")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "./output.png", "Output image file path")

	rootCmd.Execute()
}

func mergeImages(folder string, columns int, outputFile string) {
	files, err := os.ReadDir(folder)
	if err != nil {
		log.Fatalf("Failed to read folder: %v", err)
	}

	var images []image.Image
	var imgWidth, imgHeight int

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(file.Name()))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
			continue
		}
		path := filepath.Join(folder, file.Name())
		f, err := os.Open(path)
		if err != nil {
			log.Printf("Skipping %s: %v", path, err)
			continue
		}
		img, _, err := image.Decode(f)
		f.Close()
		if err != nil {
			log.Printf("Skipping %s: decode error %v", path, err)
			continue
		}
		if len(images) == 0 {
			imgWidth = img.Bounds().Dx()
			imgHeight = img.Bounds().Dy()
		} else if img.Bounds().Dx() != imgWidth || img.Bounds().Dy() != imgHeight {
			log.Fatalf("Image %s size does not match the others", file.Name())
		}
		images = append(images, img)
	}

	if len(images) == 0 {
		log.Fatal("No images loaded")
	}

	rows := (len(images) + columns - 1) / columns
	outWidth := columns * imgWidth
	outHeight := rows * imgHeight

	dst := image.NewRGBA(image.Rect(0, 0, outWidth, outHeight))
	for i, img := range images {
		x := (i % columns) * imgWidth
		y := (i / columns) * imgHeight
		draw.Draw(dst, image.Rect(x, y, x+imgWidth, y+imgHeight), img, image.Point{}, draw.Src)
	}

	out, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer out.Close()

	if strings.HasSuffix(strings.ToLower(outputFile), ".png") {
		err = png.Encode(out, dst)
	} else {
		err = jpeg.Encode(out, dst, &jpeg.Options{Quality: 90})
	}

	if err != nil {
		log.Fatalf("Failed to save image: %v", err)
	}

	fmt.Printf("Image saved to %s\n", outputFile)
}
