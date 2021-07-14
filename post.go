package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
)

// Struct to create post
type Post struct {
	ID           int
	UserName     string
	Caption      string
	ImageCaption string
	ImagePath    string
	TimeStamp    string
	AudioPath    string
	TextToSpeech string
	PostChannel  chan string
}

func newPost(userName string, caption string, imagePath string) *Post {
	return &Post{
		UserName:    userName,
		Caption:     caption,
		ImagePath:   imagePath,
		PostChannel: make(chan string)}
}

func (post *Post) imageCaption() {
	image := post.ImagePath
	var output bytes.Buffer // Buffer for stdio
	var cmd *exec.Cmd

	fmt.Println("Getting caption for the image...")

	os := runtime.GOOS
	switch os {
	case "windows":
		cmd = exec.Command("cmd", "/C", "python", "imagecap.py", image)
	case "linux":
		cmd = exec.Command("bash", "-c", "python3", "imagecap.py", image)
	default:
		cmd = exec.Command("bash", "-c", "python3", "imagecap.py", image)
	}

	// Execute the machine learning image caption script
	cmd.Stdin = strings.NewReader("From Go Take Arguments")
	cmd.Stdout = &output
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
	}

	result := output.String()
	// Store the result in the post channel
	post.PostChannel <- strings.Split(result, "Prediction Caption: ")[1]
}

func (post *Post) postTTS() {
	audio := post.AudioPath
	tts := post.TextToSpeech
	// args := []string{audio, tts}
	var output bytes.Buffer // Buffer for stdio
	var cmd *exec.Cmd

	fmt.Println("Generating Audio File For Post...")

	os := runtime.GOOS
	switch os {
	case "windows":
		cmd = exec.Command("python", "tts.py", tts, audio)
	case "linux":
		cmd = exec.Command("bash", "-c", "python3", "tts.py", tts, audio)
	default:
		cmd = exec.Command("bash", "-c", "python3", "tts.py", tts, audio)
	}

	// Execute the machine learning image caption script
	cmd.Stdin = strings.NewReader("From Go Take Arguments")
	cmd.Stdout = &output

	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
	}

	result := output.String()
	fmt.Println("--->" + result)
	// Store the result in the post channel
	// post.PostChannel <- strings.Split(result, "Prediction Caption: ")[1]
}

// Helper function to create image and return it's path
func createPostImage(r *http.Request) (string, string) {
	mf, fh, err := r.FormFile("image")

	check(err)       // If an error exists print it
	defer mf.Close() // Defer close the file

	ext := strings.Split(fh.Filename, ".")[1] // Get the extention from the fileheader

	// Check file extention if it's valid for image captioning
	// TODO
	//

	hash := sha1.New()                                       // Create a SHA1 hash
	io.Copy(hash, mf)                                        // Pass the image through the hash
	fileName := fmt.Sprintf("%x", hash.Sum(nil)) + "." + ext // Create the file name in hexadecimal form
	cwd, err := os.Getwd()                                   // Current Working Directory
	check(err)
	filePath := path.Join(cwd, "posts_images", fileName) // Create the image path
	newFile, err := os.Create(filePath)                  // Create the image file
	check(err)
	mf.Seek(0, 0)        // Reset the multipart file seeker as it was used in SHA1
	io.Copy(newFile, mf) // Copy multipart file content to new file

	return filePath, fileName
}

func createUserImage(r *http.Request) string {
	mf, fh, err := r.FormFile("image")

	check(err)       // If an error exists print it
	defer mf.Close() // Defer close the file

	ext := strings.Split(fh.Filename, ".")[1] // Get the extention from the fileheader

	hash := sha1.New()                                       // Create a SHA1 hash
	io.Copy(hash, mf)                                        // Pass the image through the hash
	fileName := fmt.Sprintf("%x", hash.Sum(nil)) + "." + ext // Create the file name in hexadecimal form
	cwd, err := os.Getwd()                                   // Current Working Directory
	check(err)
	filePath := path.Join(cwd, "users_images", fileName) // Create the image path
	newFile, err := os.Create(filePath)                  // Create the image file
	check(err)
	mf.Seek(0, 0)        // Reset the multipart file seeker as it was used in SHA1
	io.Copy(newFile, mf) // Copy multipart file content to new file

	return fileName
}
