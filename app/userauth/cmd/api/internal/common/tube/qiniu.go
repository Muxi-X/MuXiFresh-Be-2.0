package tube

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"path"
	"strings"
	"time"

	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/config"

	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
)

const (
	qnUploadTokenTTL       = 10 * 60
	qnUploadMaxFileSize    = 5 << 20
	qnUploadKeyPrefix      = "avatar/"
	qnUploadAllowedMIMEs   = "image/jpeg;image/png;image/gif;image/webp"
	qnUploadReturnBodyJSON = `{"key":"$(key)","hash":"$(etag)","mimeType":"$(mimeType)","size":$(fsize)}`
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
		AccessKey: c.Oss.AccessKey,
		SecretKey: c.Oss.SecretKey,
		Bucket:    c.Oss.BucketName,
		Domain:    c.Oss.DomainName,
	}
}

func UploadFileToQiniu(localFilePath string) (string, error) {
	mac := qbox.NewMac(Q.AccessKey, Q.SecretKey)
	cfg := storage.Config{
		Zone:          &storage.ZoneHuanan,
		UseCdnDomains: false,
		UseHTTPS:      false,
	}

	uploader := storage.NewFormUploader(&cfg)
	putPolicy := storage.PutPolicy{
		Scope: Q.Bucket,
	}
	token := putPolicy.UploadToken(mac)
	ret := storage.PutRet{}
	remoteFileName := "captcha/" + time.Now().String() + path.Base(localFilePath)
	err := uploader.PutFile(context.Background(), &ret, token, remoteFileName, localFilePath, nil)
	if err != nil {
		return "", err
	}
	return Q.Domain + "/" + ret.Key, nil
}

func GetQNToken(userID string) (string, error) {
	userID = cleanKeySegment(userID)
	if userID == "" {
		return "", errors.New("empty user id")
	}

	keyPrefix := qnUploadKeyPrefix + userID + "/"
	saveKey, err := uploadSaveKey(keyPrefix)
	if err != nil {
		return "", err
	}

	putPolicy := storage.PutPolicy{
		Scope:           Q.Bucket + ":" + keyPrefix,
		IsPrefixalScope: 1,
		Expires:         qnUploadTokenTTL,
		InsertOnly:      1,
		EndUser:         userID,
		ReturnBody:      qnUploadReturnBodyJSON,
		ForceSaveKey:    true,
		SaveKey:         saveKey,
		FsizeLimit:      qnUploadMaxFileSize,
		DetectMime:      1,
		MimeLimit:       qnUploadAllowedMIMEs,
	}
	mac := qbox.NewMac(Q.AccessKey, Q.SecretKey)
	return putPolicy.UploadToken(mac), nil
}

func uploadSaveKey(prefix string) (string, error) {
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	return prefix + time.Now().UTC().Format("20060102150405") + "-" + hex.EncodeToString(randomBytes), nil
}

func cleanKeySegment(value string) string {
	var b strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_':
			b.WriteRune(r)
		}
	}
	return b.String()
}
