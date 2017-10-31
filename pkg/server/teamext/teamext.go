package teamext

// TeamExt deal with team operations that needs integrations with App Operations (avoiding circular import)
type TeamExt interface {
	ChangeTeam(appName, teamName string) error
	ListByTeam(teamName string) ([]string, error)
}
