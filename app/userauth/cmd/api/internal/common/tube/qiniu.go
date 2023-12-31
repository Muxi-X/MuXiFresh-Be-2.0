package tube

import (
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/config"
	"context"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
	"path"
	"time"
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

func GetQNToken() string {
	var maxInt uint64 = 1 << 32
	putPolicy := storage.PutPolicy{
		Scope:   Q.Bucket,
		Expires: maxInt,
	}
	mac := qbox.NewMac(Q.AccessKey, Q.SecretKey)
	QNToken := putPolicy.UploadToken(mac)
	return QNToken
}
