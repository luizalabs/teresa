package cmd

// variables used to capture the cli flags
var (
	cfgFile         string
	debugFlag       bool
	teamIDFlag      int64
	teamNameFlag    string
	teamEmailFlag   string
	teamURLFlag     string
	appNameFlag     string
	appScaleFlag    int
	descriptionFlag string
)

const (
	deploymentSuccessMark = "----------deployment-success----------"
	deploymentErrorMark   = "----------deployment-error----------"
)
