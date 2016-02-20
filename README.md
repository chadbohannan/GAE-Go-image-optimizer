GAE Go image optimizer
======================

Go package for automatically optimizing images uploaded to Google App Engine blobstore. This reduces file size of images and, thus, reducing download times and saving you dollars.

Features:
---------
  * Files are converted to JPEG format.
  * Compression rate is changable.
    * (highly compressed) 0 --> 100 (not much compressed)
  * Resize to a maximum size in either dimension.
    * 0 = (default) unlimited / no change.
  * Returns a []byte of the compressed image.

Usage
-----
 ```go
import "github.com/chadbohannan/gaeresize"

func blobstoreUploadHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	blobs, formData, _ := blobstore.ParseUpload(r)
	
	// assume the blob is jpeg unless the client specified with a form field
	mimeType := "image/jpeg"
	if info := formData["MimeType"]; len(info) > 0 {
		mimeType = info[0]
	}

	// read the blobkey out of the parsed upload
	blobKey, _ := gaeresize.ReadBlobKey(blobs)

	// compress the blob
	params := gaeresize.NewParams(mimeType, 45, 100)
	imgBytes, _ := gaeresize.CompressBlob(c, blobKey, params)
	...
}
```
