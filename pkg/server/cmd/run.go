package cmd

import (
	"log"

	"github.com/kelseyhightower/envconfig"
	"github.com/luizalabs/teresa-api/pkg/server"
	"github.com/luizalabs/teresa-api/pkg/server/auth"
	"github.com/luizalabs/teresa-api/pkg/server/k8s"
	"github.com/luizalabs/teresa-api/pkg/server/secrets"
	"github.com/luizalabs/teresa-api/pkg/server/storage"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start Teresa gRPC server",
	Run:   runServer,
}

func init() {
	RootCmd.AddCommand(runCmd)
	runCmd.Flags().String("port", "50051", "TCP port to create a listener")
}

func runServer(cmd *cobra.Command, args []string) {
	port, _ := cmd.Flags().GetString("port")
	db, err := getDB()
	if err != nil {
		log.Fatal("Error on connect to Database: ", err)
	}

	st, err := getStorage()
	if err != nil {
		log.Fatal("Error configuring storage: ", err)
	}

	k8s, err := getK8s()
	if err != nil {
		log.Fatal("Error configuring k8s client: ", err)
	}

	sec, err := getSecrets()
	if err != nil {
		log.Fatal("Error on get secrects information ", err)
	}

	a, err := getAuth(sec)
	if err != nil {
		log.Fatal("Error on get auth information ", err)
	}

	tlsCert, err := sec.TLSCertificate()
	if err != nil {
		log.Fatal("Error on get TLS cert ", err)
	}
	s, err := server.New(server.Options{
		Port:    port,
		Auth:    a,
		DB:      db,
		TLSCert: tlsCert,
		Storage: st,
		K8s:     k8s,
	})
	if err != nil {
		log.Fatal("Error on create Server: ", err)
	}

	log.Println("Start TERESA server on port ", port)
	log.Print(s.Run())
}

func getSecrets() (secrets.Secrets, error) {
	conf := new(secrets.FileSystemSecretsConfig)
	if err := envconfig.Process("teresasecrets", conf); err != nil {
		return nil, err
	}
	return secrets.NewFileSystemSecrets(*conf)
}

func getAuth(s secrets.Secrets) (auth.Auth, error) {
	private, err := s.PrivateKey()
	if err != nil {
		return nil, err
	}
	public, err := s.PublicKey()
	if err != nil {
		return nil, err
	}
	return auth.New(private, public), nil
}

func getStorage() (storage.Storage, error) {
	conf := new(storage.Config)
	if err := envconfig.Process("teresafilestorage", conf); err != nil {
		return nil, err
	}
	return storage.New(conf)
}

func getK8s() (k8s.Client, error) {
	conf := new(k8s.Config)
	if err := envconfig.Process("teresak8s", conf); err != nil {
		return nil, err
	}
	return k8s.New(conf)
}
