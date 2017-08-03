package storage

import "strings"

type Minio struct {
	Storage
}

func (m *Minio) Type() string {
	return string(MinioType)
}

func (m *Minio) PodEnvVars() map[string]string {
	s3 := m.Storage.(*S3)
	endpoint := strings.Split(trimEndpoint(s3.Endpoint), ":")
	return map[string]string{
		"S3_HOST":      endpoint[0],
		"S3_PORT":      endpoint[1],
		"MINIO_BUCKET": s3.Bucket,
	}
}

func trimEndpoint(endpoint string) string {
	for _, s := range []string{"http://", "https://"} {
		endpoint = strings.TrimPrefix(endpoint, s)
	}
	return endpoint
}

func newMinio(conf *Config) Storage {
	s3 := newS3(conf)
	return &Minio{Storage: s3}
}
