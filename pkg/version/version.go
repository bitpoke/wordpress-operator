package version

var (
	AppGitState  = ""
	AppGitCommit = ""
	AppVersion   = "canary"
)

func Get() string {
	return AppVersion
}
