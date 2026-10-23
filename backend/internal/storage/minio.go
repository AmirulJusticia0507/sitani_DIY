package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinioClient adalah Client nyata di atas minio-go.
type MinioClient struct {
	// pointer supaya niatnya jelas: objek internal dibuat sekali saat init.
	minio  *minio.Client
	bucket string
}

// NewMinioClient membuka koneksi ke MinIO/S3, memastikan bucket tersedia
// (dibuat jika belum ada), dan mengembalikan Client. Kegagalan koneksi
// maupun pembuatan bucket dikembalikan sebagai error.
func NewMinioClient(ctx context.Context, cfg Config) (*MinioClient, error) {
	if !cfg.Siap() {
		return nil, ErrStorageTidakSiap
	}
	// format endpoint: host:port tanpa skema (https dipilih lewat MINIO_SECURE)
	endpoint := strings.TrimSpace(cfg.Endpoint)
	endpoint = strings.TrimPrefix(endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")

	klien, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AksesKunci, cfg.Rahasia, ""),
		Secure: cfg.HTTPS,
	})
	if err != nil {
		return nil, fmt.Errorf("buat klien object storage: %w", err)
	}

	ada, err := klien.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("periksa bucket %s: %w", cfg.Bucket, err)
	}
	if !ada {
		if err := klien.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("buat bucket %s: %w", cfg.Bucket, err)
		}
	}
	return &MinioClient{minio: klien, bucket: cfg.Bucket}, nil
}

// Unggah menyimpan berkas dengan metadata tipe konten.
func (m *MinioClient) Unggah(ctx context.Context, objectKey string, reader io.Reader, ukuran int64, jenis string) error {
	_, err := m.minio.PutObject(ctx, m.bucket, objectKey, reader, ukuran, minio.PutObjectOptions{
		ContentType: jenis,
	})
	if err != nil {
		return fmt.Errorf("unggah %s: %w", objectKey, err)
	}
	return nil
}

// URLUnduh menerbitkan URL presign GET berumur pendek; kueri respons author
// identitas diberkas server respons.
func (m *MinioClient) URLUnduh(ctx context.Context, objectKey string) (string, error) {
	u, err := m.minio.PresignedGetObject(ctx, m.bucket, objectKey, preSignDurasi, url.Values{})
	if err != nil {
		return "", fmt.Errorf("terbitkan URL unduh %s: %w", objectKey, err)
	}
	return u.String(), nil
}

// Periksa memastikan bucket dapat diakses.
func (m *MinioClient) Periksa(ctx context.Context) error {
	ada, err := m.minio.BucketExists(ctx, m.bucket)
	if err != nil {
		return fmt.Errorf("periksa bucket %s: %w", m.bucket, err)
	}
	if !ada {
		return fmt.Errorf("bucket %s tidak tersedia", m.bucket)
	}
	return nil
}
