package cmd

import (
	"crypto/tls"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/kelseyhightower/envconfig"
	"github.com/luizalabs/teresa/pkg/server"
	"github.com/luizalabs/teresa/pkg/server/auth"
	"github.com/luizalabs/teresa/pkg/server/deploy"
	"github.com/luizalabs/teresa/pkg/server/k8s"
	"github.com/luizalabs/teresa/pkg/server/resource"
	"github.com/luizalabs/teresa/pkg/server/secrets"
	"github.com/luizalabs/teresa/pkg/server/storage"
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
	runCmd.Flags().Bool("tls", false, "enable TLS")
	runCmd.Flags().Bool("debug", false, "enable debug mode")
}

func runServer(cmd *cobra.Command, args []string) {
	port, err := cmd.Flags().GetString("port")
	if err != nil {
		log.WithError(err).Fatal("invalid port parameter")
	}

	useTLS, err := cmd.Flags().GetBool("tls")
	if err != nil {
		log.WithError(err).Fatal("invalid tls parameter")
	}

	debug, err := cmd.Flags().GetBool("debug")
	if err != nil {
		log.WithError(err).Fatal("invalid debug parameter")
	}

	db, err := getDB()
	if err != nil {
		log.WithError(err).Fatal("failed to connect to database")
	}

	st, err := getStorage()
	if err != nil {
		log.WithError(err).Fatal("failed to configure storage")
	}

	k8s, err := getK8s()
	if err != nil {
		log.WithError(err).Fatal("failed to configure k8s client")
	}

	sec, err := getSecrets()
	if err != nil {
		log.WithError(err).Fatal("failed to get secrets data")
	}

	a, err := getAuth(sec)
	if err != nil {
		log.WithError(err).Fatal("failed to get auth data")
	}

	var tlsCert *tls.Certificate
	if useTLS {
		tlsCert, err = sec.TLSCertificate()
		if err != nil {
			log.WithError(err).Fatal("failed to get TLS cert")
		}
	}

	deployOpt, err := getDeployOpt()
	if err != nil {
		log.Fatal("Error getting deploy configuration:", err)
	}

	tmpl, err := getResourceTemplater()
	if err != nil {
		log.WithError(err).Fatal("failed to configure resource templater")
	}

	s, err := server.New(server.Options{
		Port:      port,
		Auth:      a,
		DB:        db,
		TLSCert:   tlsCert,
		Storage:   st,
		K8s:       k8s,
		Tmpl:      tmpl,
		Exe:       resource.NewTemplateExecuter(),
		DeployOpt: deployOpt,
		Debug:     debug,
	})
	if err != nil {
		log.WithError(err).Fatal("failed to create server")
	}

	log.Info("starting teresa-server on port ", port)
	log.Info(s.Run())
}

func getSecrets() (secrets.Secrets, error) {
	conf := new(secrets.FileSystemSecretsConfig)
	if err := envconfig.Process("teresa_secrets", conf); err != nil {
		return nil, err
	}
	return secrets.NewFileSystemSecrets(conf)
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
	if err := envconfig.Process("teresa_storage", conf); err != nil {
		return nil, err
	}
	return storage.New(conf)
}

func getK8s() (k8s.Client, error) {
	conf := new(k8s.Config)
	if err := envconfig.Process("teresa_k8s", conf); err != nil {
		return nil, err
	}
	return k8s.New(conf)
}

func getDeployOpt() (*deploy.Options, error) {
	conf := new(deploy.Options)
	if err := envconfig.Process("teresa_deploy", conf); err != nil {
		return nil, err
	}
	return conf, nil
}

func getResourceTemplater() (resource.Templater, error) {
	cfg := new(resource.Config)
	if err := envconfig.Process("teresa_resource", cfg); err != nil {
		return nil, err
	}
	return resource.NewTemplater(cfg, http.DefaultClient), nil
}
