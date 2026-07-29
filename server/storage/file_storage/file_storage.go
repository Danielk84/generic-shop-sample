package file_storage

import (
	"context"
	"errors"
	"fmt"
	"generic-shop-sample/internal/config"
	"generic-shop-sample/internal/logger"
	"io"
	"mime/multipart"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsC "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/gabriel-vasile/mimetype"
)

var (
	ErrInvalidMimeType   = errors.New("invalid mime type")
	ErrInvalidBucketName = errors.New("invalid bucket name")
)

type FileStoreClient = *s3.Client

func NewFileStoreClient(ctx context.Context, config config.AwsS3Config) (FileStoreClient, error) {
	cred := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(config.Key, config.Secret, ""))
	cfg, err := awsC.LoadDefaultConfig(ctx,
		awsC.WithRegion(config.Region),
		awsC.WithCredentialsProvider(cred),
	)
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(config.Endpoint)
		o.UsePathStyle = true
	})
	return client, nil
}

type FileStore interface {
	Upload(ctx context.Context, fh *multipart.FileHeader, key string) (string, error)
	Download(ctx context.Context, bucket, key string) (*s3.GetObjectOutput, error)
	Delete(ctx context.Context, key string) error
	BulkDelete(ctx context.Context, keys []string) error
}

func NewFileStore(ctx context.Context, config config.FileStorageConfig, client FileStoreClient, bucket string) FileStore {
	log := logger.GetLogger()
	f := &fileManager{
		log:    log,
		config: config,
		bucket: bucket,
		client: client,
	}
	if err := f.createBucket(ctx); err != nil {
		log.Debug("NewFileStore", "error", err)
		panic(err)
	}
	return f
}

type fileManager struct {
	log    logger.Logger
	config config.FileStorageConfig
	bucket string
	client FileStoreClient
}

func (f *fileManager) isBucketExists(ctx context.Context) (bool, error) {
	if f.bucket == "" {
		return false, ErrInvalidBucketName
	}
	_, err := f.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(f.bucket),
	})
	if err == nil {
		return true, nil
	}
	var apiErr smithy.APIError
	if !errors.As(err, &apiErr) {
		return false, err
	}
	switch apiErr.(type) {
	case *types.NotFound:
		return false, nil
	default:
		return true, err
	}
}

func (f *fileManager) createBucket(ctx context.Context) error {
	exists, err := f.isBucketExists(ctx)
	if err != nil {
		f.log.Warn("fileManager.createBucket", "error", err)
		return err
	}
	if exists {
		return nil
	}
	_, err = f.client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(f.bucket),
	})
	if err != nil {
		f.log.Debug("fileManager.createBucket", "error", err)
	}
	return err
}

func (f *fileManager) validate(file multipart.File) (string, error) {
	buf := make([]byte, 512)
	n, err := file.ReadAt(buf, io.SeekStart)
	if err != nil {
		return "", err
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	mtype := mimetype.Detect(buf[:n])
	if mtypeStr := mtype.String(); !mimetype.EqualsAny(mtypeStr, f.config.AllowedImgMimetype...) {
		f.log.Debug("fileManager.Validate", "mimeType", mtypeStr, "expected", f.config.AllowedImgMimetype)
		return "", ErrInvalidMimeType
	}
	return mtype.Extension(), nil
}

func (f *fileManager) Upload(ctx context.Context, fh *multipart.FileHeader, key string) (string, error) {
	file, err := fh.Open()
	if err != nil {
		f.log.Warn("fileManager.Upload", "error", err)
		return "", err
	}
	defer func() { _ = file.Close() }()

	fileExt, err := f.validate(file)
	if err != nil {
		f.log.Debug("fileManager.Upload", "error", err)
		return "", err
	}

	fileKey := fmt.Sprintf("%s%s", key, fileExt)
	_, err = f.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(f.bucket),
		Key:    aws.String(fileKey),
		Body:   file,
	})
	if err != nil {
		f.log.Error("fileManager.Upload", "error", err)
	}
	return fmt.Sprintf("%s/%s", f.bucket, fileKey), nil
}

func (f *fileManager) Download(ctx context.Context, bucket, key string) (obj *s3.GetObjectOutput, err error) {
	obj, err = f.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		f.log.Debug("fileManager.Download", "error", err)
	}
	return
}

func (f *fileManager) Delete(ctx context.Context, key string) error {
	fileKey := strings.TrimLeft(key, f.bucket+"/")
	_, err := f.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(f.bucket),
		Key:    aws.String(fileKey),
	})
	if err != nil {
		var apiErr smithy.APIError
		if !errors.As(err, &apiErr) {
			f.log.Error("fileManager.Delete", "error", err)
			return err
		}
		switch apiErr.(type) {
		case *types.NotFound:
			return nil
		default:
			f.log.Error("fileManager.Delete", "error", err)
			return err
		}
	}
	return nil
}

func (f *fileManager) BulkDelete(ctx context.Context, keys []string) error {
	objs := make([]types.ObjectIdentifier, 0, len(keys))
	for _, k := range keys {
		fileKey := strings.TrimLeft(k, f.bucket+"/")
		objs = append(objs, types.ObjectIdentifier{
			Key: aws.String(fileKey),
		})
	}
	_, err := f.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(f.bucket),
		Delete: &types.Delete{
			Objects: objs,
			Quiet:   aws.Bool(true),
		},
	})
	if err != nil {
		f.log.Error("fileManager.BulkDelete",
			"error", err,
			"bucket", f.bucket,
			"keys", keys)
	}
	return err
}
