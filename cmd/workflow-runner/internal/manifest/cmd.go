package manifest

var (
	cmdCalcScore        []string
	cmdUpdateDist       []string
	cmdUpdateGitMetrics []string
	cmdUpdateGitPartial []string
	cmdUpdateDepsDev    []string
	cmdSyncGitMetrics   []string
	cmdEnumPlatforms    []string
)

var (
	binCalcScore        = "scores-calculator"
	binUpdateDist       = "dist-packages-collector"
	binUpdateGitMetrics = "git-metrics-collector"
	binUpdateDepsDev    = "deps-dev-collector"
	binSyncGitMetrics   = "git-metrics-sync"
	binEnumPlatforms    = "git-platforms-enumerator"
)

func initCmds() {
	dir := "./bin/"

	commonArgs := []string{"-c", "config.json"}

	cmdCalcScore = append([]string{dir + binCalcScore}, commonArgs...)
	cmdUpdateDist = append([]string{dir + binUpdateDist}, commonArgs...)
	cmdUpdateGitMetrics = append([]string{dir + binUpdateGitMetrics}, append(commonArgs, "--force-update-all")...)
	cmdUpdateGitPartial = append([]string{dir + binUpdateGitMetrics}, commonArgs...)
	cmdUpdateDepsDev = append([]string{dir + binUpdateDepsDev}, commonArgs...)
	cmdSyncGitMetrics = append([]string{dir + binSyncGitMetrics}, commonArgs...)
	cmdEnumPlatforms = append([]string{dir + binEnumPlatforms}, commonArgs...)

}
