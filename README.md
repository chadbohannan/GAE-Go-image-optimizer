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
  * Returns a key to a new blobstore entity.

Usage
-----
  ```go
    import "github.com/chadbohannan/gae-go-image-optimizer"
    
    func blobstoreUploadHandler(w http.ResponseWriter, r *http.Request) {
      c := appengine.NewContext(r)

      // Get the original file
	  blobKey := gaeresize.ProcessRequest(r, nil)

      // Create options, quality = 75, size = 1600
      o := optimg.NewDefaultOptions(c, 75, 1600)

      // Get an optimized blob key
      compressedBlobKey := gaeresize.ProcessRequest(r, o)

      ...
    }
  ```
