package cmd

var (
	fURL           = "url"
	fUsername      = "username"
	fPassword      = "password"
	fVerbose       = "verbose"
	fSkipSSLVerify = "skipSSLVerify"
)

func init() {
	rootCmd.PersistentFlags().String(fURL, "https://localhost:9070/api", "Base URL of VTM")
	rootCmd.PersistentFlags().StringP(fUsername, "u", "admin", "username w/ REST API access")
	rootCmd.PersistentFlags().StringP(fPassword, "p", "admin", "password of username")
	rootCmd.PersistentFlags().BoolP(fVerbose, "v", false, "Be verbose")
	rootCmd.PersistentFlags().BoolP(fSkipSSLVerify, "k", false, "Skip SSL certificate verification")
}
