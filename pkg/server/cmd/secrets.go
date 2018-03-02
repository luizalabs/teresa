package cmd

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/spf13/cobra"
)

var replaceStorageSecretCmd = &cobra.Command{
	Use:   "replace-storage-secret",
	Short: "Replace the storage secret for all apps",
	Run:   replaceStorageSecret,
}

func init() {
	RootCmd.AddCommand(replaceStorageSecretCmd)
	replaceStorageSecretCmd.Flags().String("id", "", "key identity")
	replaceStorageSecretCmd.Flags().String("key", "", "secret access key")
}

func replaceStorageSecret(cmd *cobra.Command, args []string) {
	id, err := cmd.Flags().GetString("id")
	if err != nil || id == "" {
		log.WithError(err).Fatal("invalid id parameter")
	}
	key, err := cmd.Flags().GetString("key")
	if err != nil || key == "" {
		log.WithError(err).Fatal("invalid key parameter")
	}

	k8s, err := getK8s()
	if err != nil {
		log.WithError(err).Fatal("can't create k8s client")
	}
	st, err := getStorage()
	if err != nil {
		log.WithError(err).Fatal("can't create storage client")
	}
	data := st.AccessData()
	data["accesskey"] = []byte(id)
	data["secretkey"] = []byte(key)
	secretName := st.K8sSecretName()

	apps, err := k8s.NamespaceListByLabel(app.TeresaTeamLabel, "")
	if err != nil {
		log.WithError(err).Fatal("can't get app list")
	}
	for _, app := range apps {
		if err := k8s.CreateOrUpdateSecret(string(app), secretName, data); err != nil {
			log.WithError(err).Fatalf("can't update secret for app %s", app)
		}
	}

	fmt.Println("Storage secrets replaced")
}
