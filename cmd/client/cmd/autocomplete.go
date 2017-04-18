package cmd

import "github.com/spf13/cobra"

// from hugo
var autocompleteCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate bash autocompletion script for Teresa cli",
	Long: `Generate bash autocompletion script for Teresa cli
	By default, the file is written directly to /etc/bash_completion.d
for convenience, and the command may need superuser rights, e.g.:

    $ sudo teresa gen autocomplete

Add ` + "`--completionfile=/path/to/file`" + ` flag to set alternative
file-path and name.

Logout and in again to reload the completion scripts,
or just source them in directly:

    $ . /etc/bash_completion`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Root().GenBashCompletionFile(autocompleteTarget); err != nil {
			log.Warning("If you runned as non-root without a --completionfile='', we may have had permission issues.")
			log.Fatal(err)
		}
		log.Infof("Bash completion file saved to %s\n", autocompleteTarget)
	},
}

func init() {
	autocompleteCmd.PersistentFlags().StringVarP(&autocompleteTarget, "completionfile", "", "/etc/bash_completion.d/teresa.sh", "Autocompletion file")
	autocompleteCmd.PersistentFlags().SetAnnotation("completionfile", cobra.BashCompFilenameExt, []string{})
	RootCmd.AddCommand(autocompleteCmd)
}
