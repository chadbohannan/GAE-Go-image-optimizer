/***************************************************************
*
*   GAE Go automatic blob image optimizer
*
*   Created by Tomi Hiltunen 2013.
*   http://www.linkedin.com/in/tomihiltunen
*
*   https://github.com/TomiHiltunen/GAE-Go-image-optimizer
*
*       - Use this script however you wish.
*       - Do not remove any copyrights/comments on any files included.
*       - All use is on your own risk.
*
*   Intented use:
*       - Drop-in replacement for GAE's blobstore.ParseUploads(...)
*       - Automatically optimized any images uploaded
*         to Google App Engine blobstore.
*           - Reduces data amount in the blobstore.
*           - Reduces download times.
*
*   Adulterated use:
*       - Forked the lib to refactor into a nondestructive library call
*       - Compressed artifacts are small enough to store in datastore
*       - Artifacts in datastore can most easily be served securly
***************************************************************/
package gaeresize

import (
	"appengine"
	"appengine/blobstore"
	"bytes"
	"errors"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"math"
	"strings"
)

//  Allowed mime-types; should be the ones supported by Go image package.
var (
	allowedMimeTypes = map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
	}
)

type Params struct {
	MimeType string
	Quality  int // The quality of the JPEG output (0-100)
	Size     int // Maximum dimension (width/height) for the photo
}

func NewDefaultOptions(mimeType string) *Params {
	return &Params{
		MimeType: mimeType, // required to be in the set of supported types
		Quality:  75,       // 75 is highly compressed but not visually noticable.
		Size:     0,        // 0 = do not resize, otherwise this is the maximum dimension
	}
}

func NewParams(mimeType string, quality, size int) *Params {
	return &Params{
		MimeType: mimeType,
		Quality:  quality,
		Size:     size,
	}
}

// utility function to get a blobkey from the mess that blobstore parses out
func ReadBlobKey(blobs map[string][]*blobstore.BlobInfo) (appengine.BlobKey, error) {
	// read the file metadata from the blobs
	if file := blobs["file"]; len(file) == 0 {
		return "", errors.New("Missing file data.")
	} else {
		return file[0].BlobKey, nil
	}
}

// - Reads a blobstore key, writes to a cloud storage bucket file.
// - Returns a []byte of a JPEG or a non-nil error
func CompressBlob(c appengine.Context, blobKey appengine.BlobKey, params *Params) ([]byte, error) {
	// Check that the blob is of supported mime-type
	if !allowedMimeTypes[strings.ToLower(params.MimeType)] {
		return nil, errors.New("Unsupported mime-type:" + params.MimeType)
	}

	// Instantiate blobstore reader
	reader := blobstore.NewReader(c, blobKey)

	// Instantiate the image object
	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}
	// Resize if necessary an maintain aspect ratio
	if params.Size > 0 && (img.Bounds().Max.X > params.Size || img.Bounds().Max.Y > params.Size) {
		size_x := img.Bounds().Max.X
		size_y := img.Bounds().Max.Y
		if size_x > params.Size {
			size_x_before := size_x
			size_x = params.Size
			size_y = int(math.Floor(float64(size_y) * float64(float64(size_x)/float64(size_x_before))))
		}
		if size_y > params.Size {
			size_y_before := size_y
			size_y = params.Size
			size_x = int(math.Floor(float64(size_x) * float64(float64(size_y)/float64(size_y_before))))
		}
		img = Resize(img, img.Bounds(), size_x, size_y)
	}

	// Write JPEG to buffer
	b := new(bytes.Buffer)
	o := &jpeg.Options{Quality: params.Quality}
	if err := jpeg.Encode(b, img, o); err != nil {
		return nil, err
	}

	// Return image content
	return b.Bytes(), nil
}
