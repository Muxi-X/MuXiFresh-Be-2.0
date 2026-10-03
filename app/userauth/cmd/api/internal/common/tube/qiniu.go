package tube

import (
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/config"
	"context"
	"fmt"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
	"path/filepath"
	"strings"
	"time"
)

const (
	// Keep upload credentials short-lived. The Qiniu SDK treats Expires as a
	// duration in seconds and converts it to an absolute deadline when signing.
	uploadTokenExpires uint64 = 5 * 60

	// Server-side captcha uploads remain image-only. Use an explicit raster-image
	// allowlist so active content such as HTML and SVG cannot be hosted on the
	// application's trusted storage domain.
	allowedImageMimeTypes = "image/jpeg;image/png;image/gif;image/webp;image/bmp;image/avif;image/heic;image/heif"

	// Client uploads also support passive document attachments and ZIP archives.
	// Keep this as an explicit allowlist so other executable or active content
	// types remain blocked.
	allowedDocumentMimeTypes = "application/pdf;application/vnd.openxmlformats-officedocument.wordprocessingml.document;application/zip"
	allowedClientUploadMimes = allowedImageMimeTypes + ";" + allowedDocumentMimeTypes
	maxUploadSize            = int64(10 << 20)

	// Client upload keys are scoped under this per-user prefix so a token cannot
	// write elsewhere in the bucket. Clients must generate keys under this
	// prefix (e.g. avatar/<userId>/<fileName>).
	clientUploadKeyPrefix = "avatar/"
)

type Qiniu struct {
	AccessKey string
	SecretKey string
	Bucket    string
	Domain    string
}

var Q Qiniu

func Load(c config.Config) {
	Q = Qiniu{
		AccessKey: c.Infra.ObjectStorage.AccessKey,
		SecretKey: c.Infra.ObjectStorage.SecretKey,
		Bucket:    c.Infra.ObjectStorage.BucketName,
		Domain:    c.Infra.ObjectStorage.DomainName,
	}
}

func UploadFileToQiniu(localFilePath string) (string, error) {
	mac := qbox.NewMac(Q.AccessKey, Q.SecretKey)
	cfg := storage.Config{
		Zone:          &storage.ZoneHuanan,
		UseCdnDomains: false,
		UseHTTPS:      true,
	}

	uploader := storage.NewFormUploader(&cfg)
	remoteFileName := fmt.Sprintf("captcha/%d-%s", time.Now().UnixNano(), filepath.Base(localFilePath))
	putPolicy := newServerImageUploadPolicy(Q.Bucket, remoteFileName)
	token := putPolicy.UploadToken(mac)
	ret := storage.PutRet{}
	err := uploader.PutFile(context.Background(), &ret, token, remoteFileName, localFilePath, nil)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(Q.Domain, "/") + "/" + ret.Key, nil
}

func GetQNToken(endUser string) string {
	putPolicy := newClientUploadPolicy(Q.Bucket, endUser)
	mac := qbox.NewMac(Q.AccessKey, Q.SecretKey)
	return putPolicy.UploadToken(mac)
}

// newClientUploadPolicy scopes the token to the caller's directory
// (bucket:avatar/<endUser>/) via IsPrefixalScope, so a leaked token cannot
// write outside that prefix. It keeps content detection, the
// image/PDF/DOCX/ZIP allowlist, a size limit, short expiry and insert-only
// semantics.
func newClientUploadPolicy(bucket, endUser string) storage.PutPolicy {
	return storage.PutPolicy{
		Scope:           bucket + ":" + clientUploadKeyPrefix + endUser + "/",
		IsPrefixalScope: 1,
		Expires:         uploadTokenExpires,
		InsertOnly:      1,
		EndUser:         endUser,
		FsizeLimit:      maxUploadSize,
		DetectMime:      1,
		MimeLimit:       allowedClientUploadMimes,
	}
}

// Server-side uploads know the destination key in advance, so grant access to
// that exact object rather than the whole bucket.
func newServerImageUploadPolicy(bucket, key string) storage.PutPolicy {
	return storage.PutPolicy{
		Scope:        bucket + ":" + key,
		Expires:      uploadTokenExpires,
		InsertOnly:   1,
		ForceSaveKey: true,
		SaveKey:      key,
		FsizeLimit:   maxUploadSize,
		DetectMime:   1,
		MimeLimit:    allowedImageMimeTypes,
	}
}
