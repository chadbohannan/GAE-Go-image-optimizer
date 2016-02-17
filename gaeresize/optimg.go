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
*       - Forked the lib to transform into simple, nondestructive library call
***************************************************************/
package gaeresize

import (
	"appengine"
	"appengine/blobstore"
	"errors"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"math"
	"net/http"
	"net/url"
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

type Options struct {
	Quality int // The quality of the JPEG output (0-100)
	Size    int // Maximum dimension (width/height) for the photo
	Context appengine.Context // AppEngine context    
}

func NewDefaultOptions(c appengine.Context) *Options {
	return &Options{
		Context: c,
		Quality: 75, // 75 is highly compressed but not visually noticable.
		Size:    0,  // 0 = do not resize, otherwise this is the maximum dimension
	}
}

func NewOptions(c appengine.Context, quality, size, int) *Options {
	return &Options{
		Context: c,
		Quality: quality,
		Size:    size,
	}
}

// r       - a request from the appengine/blobstore upload service callback
// options - compression resize options. nil to return the original blob key
// returns - struct url.Values map[string][]string, error
func ProcessRequest(r *http.Request, options *Options) (string, error) {
	blobs, values, err = blobstore.ParseUpload(r)
	
	// read the file metadata from the blobstore
	file := blobs["file"]
	if len(file) == 0 {
		return "", errors.New("Missing file data.")
	}
	blobKey := file[0].BlobKey
	if options == nil {
		return blobKey
	}
	mimeType := strings.ToLower(blob.ContentType)
	return ProcessBlob(blobKey, mimeType, options)
}


// - Only supported image types will be processed. Others will be returned as-is.
// - Resizes the image if necessary.
// - Writes the new compressed JPEG to blobstore.
// - Deletes the old blob and substitutes the old BlobInfo with the new one.
func ProcessBlob(blobKey, mimeType string, options *Options) (string, error) {
	// Check that the blob is of supported mime-type
	if !allowedMimeTypes[strings.ToLower(mimeType)] {
		return "", errors.New("Unsupported mime-type.")
	}
	// Instantiate blobstore reader
	reader := blobstore.NewReader(options.Context, blobKey)
	
	// Instantiate the image object
	img, _, err := image.Decode(reader)
	if err != nil {
		return
	}
	// Resize if necessary
	// Maintain aspect ratio!
	if options.Size > 0 && (img.Bounds().Max.X > options.Size || img.Bounds().Max.Y > options.Size) {
		size_x := img.Bounds().Max.X
		size_y := img.Bounds().Max.Y
		if size_x > options.Size {
			size_x_before := size_x
			size_x = options.Size
			size_y = int(math.Floor(float64(size_y) * float64(float64(size_x)/float64(size_x_before))))
		}
		if size_y > options.Size {
			size_y_before := size_y
			size_y = options.Size
			size_x = int(math.Floor(float64(size_x) * float64(float64(size_y)/float64(size_y_before))))
		}
		img = resize.Resize(img, img.Bounds(), size_x, size_y)
	}
	// JPEG options
	o := &jpeg.Options{
		Quality: options.Quality,
	}
	// Open writer
	writer, err := blobstore.Create(options.Context, "image/jpeg")
	if err != nil {
		return "", err
	}
	// Write to blobstore
	if err := jpeg.Encode(writer, img, o); err != nil {
		_ = writer.Close()
		return "", err
	}
	// Close writer
	if err := writer.Close(); err != nil {
		return "", err
	}
	// Get key
	newKey, err := writer.Key()
	if err != nil {
		return "", err
	}
	// Get new BlobInfo
	newBlobInfo, err := blobstore.Stat(options.Context, newKey)
	if err != nil {
		return "", err
	}
	return newBlobInfo.BlobKey, nil
}
