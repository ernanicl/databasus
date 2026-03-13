package config

type parsedFlags struct {
	host  *string
	dbID  *string
	token *string

	sources map[string]string
}
