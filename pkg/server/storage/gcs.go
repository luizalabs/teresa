package storage

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	gcsst "cloud.google.com/go/storage"
	"golang.org/x/net/context"
)

const GoogleCredentialsEnv = "GOOGLE_APPLICATION_CREDENTIALS"

var ErrInvalidCred = fmt.Errorf("Invalid Storage Credentials")

type GcsClient interface {
	PutObject(bucket, path string, file io.ReadSeeker) error
}

type concretGcsClient struct {
	client *gcsst.Client
}

type Gcs struct {
	Bucket string
	Cred   string
	Client GcsClient
}

func (*Gcs) K8sSecretName() string {
	return "gcs-storage"
}

func (g *Gcs) AccessData() map[string][]byte {
	return map[string][]byte{
		"builder-bucket": []byte(g.Bucket),
		"key.json":       []byte(g.Cred),
	}
}

func (g *Gcs) UploadFile(path string, file io.ReadSeeker) error {
	return g.Client.PutObject(g.Bucket, path, file)
}

func (*Gcs) Type() string {
	return string(GcsType)
}

func (*Gcs) PodEnvVars() map[string]string {
	return make(map[string]string)
}

func newGcs(conf *Config) (Storage, error) {
	client, err := newConcretGcsClient()
	if err != nil {
		return nil, err
	}

	gcpCred, err := getCredString()
	if err != nil {
		return nil, err
	}

	return &Gcs{
		Client: client,
		Bucket: conf.AwsBucket, //TODO: please generalize me
		Cred:   gcpCred,
	}, nil
}

func getCredString() (string, error) {
	credPath := os.Getenv(GoogleCredentialsEnv)
	if credPath == "" {
		return "", ErrInvalidCred
	}
	b, err := ioutil.ReadFile(credPath)
	if err != nil {
		return "", ErrInvalidCred
	}
	return string(b), nil
}

func (c *concretGcsClient) PutObject(bucket, path string, file io.ReadSeeker) error {
	w := c.client.Bucket(bucket).Object(path).NewWriter(context.Background())
	if _, err := io.Copy(w, file); err != nil {
		return err
	}
	return w.Close()
}

func newConcretGcsClient() (GcsClient, error) {
	client, err := gcsst.NewClient(context.Background())
	if err != nil {
		return nil, err
	}
	return &concretGcsClient{client}, nil
}
