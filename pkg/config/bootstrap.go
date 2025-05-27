package config

import (
	"os"
	"reflect"
	"unsafe"

	"github.com/HUSTSecLab/criticality_score/pkg/logger"
	"github.com/HUSTSecLab/criticality_score/pkg/storage"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	databaseRegisted = false
	logRegisted      = false
)

func RegistConfigFileFlags(flag *pflag.FlagSet) {
	flag.StringP("config", "c", "", "app config file, in json or yaml format,\ncan set by environment APP_CONFIG_FILE")
	viper.BindEnv("config-file", "APP_CONFIG_FILE")
	viper.BindPFlag("config-file", flag.Lookup("config"))
}

func RegistDatabaseFlags(flag *pflag.FlagSet) {
	databaseRegisted = true
	flag.String("db-host", "", "database host,\ncan set by environment DB_HOST")
	flag.String("db-port", "", "database port,\ncan set by environment DB_PORT")
	flag.String("db-user", "", "database user,\ncan set by environment DB_USER")
	flag.String("db-password", "", "database password,\ncan set by environment DB_PASSWORD")
	flag.String("db-password-file", "", "database password file, if db-password is set, this will be ignored,\ncan set by environment DB_PASSWORD_FILE")
	flag.Bool("db-use-ssl", false, "use ssl to connect database,\ncan set by environment DB_USE_SSL")

	viper.BindPFlag("db.host", flag.Lookup("db-host"))
	viper.BindPFlag("db.port", flag.Lookup("db-port"))
	viper.BindPFlag("db.user", flag.Lookup("db-user"))
	viper.BindPFlag("db.password", flag.Lookup("db-password"))
	viper.BindPFlag("db.password-file", flag.Lookup("db-password-file"))
	viper.BindPFlag("db.use-ssl", flag.Lookup("db-use-ssl"))

	viper.BindEnv("db.host", "DB_HOST")
	viper.BindEnv("db.port", "DB_PORT")
	viper.BindEnv("db.user", "DB_USER")
	viper.BindEnv("db.password", "DB_PASSWORD")
	viper.BindEnv("db.password-file", "DB_PASSWORD_FILE")
	viper.BindEnv("db.use-ssl", "DB_USE_SSL")
}

func RegistLogFlags(flag *pflag.FlagSet) {
	logRegisted = true
	flag.StringP("log-level", "v", "info", "increase log verbosity,\ncan set by environment LOG_LEVEL")
	flag.String("log-type", "console", "log type: console, file, es,\ncan set by environment LOG_TYPE")
	flag.String("log-path", "", "log path, only used when log-type is file,\ncan set by environment LOG_PATH")
	flag.String("log-es-url", "", "elasticsearch url, only used when log-type is es,\ncan set by environment LOG_ES_URL")
	flag.String("log-es-index", "", "elasticsearch index, only used when log-type is es,\ncan set by environment LOG_ES_INDEX")

	viper.BindPFlag("log.level", flag.Lookup("log-level"))
	viper.BindPFlag("log.type", flag.Lookup("log-type"))
	viper.BindPFlag("log.path", flag.Lookup("log-path"))
	viper.BindPFlag("log.es-url", flag.Lookup("log-es-url"))
	viper.BindPFlag("log.es-index", flag.Lookup("log-es-index"))
	viper.BindPFlag("log.es-user", flag.Lookup("log-es-user"))
	viper.BindPFlag("log.es-password", flag.Lookup("log-es-password"))
	viper.BindPFlag("log.es-cert", flag.Lookup("log-es-cert"))

	viper.BindEnv("log.debug", "LOG_DEBUG")
	viper.BindEnv("log.type", "LOG_TYPE")
	viper.BindEnv("log.path", "LOG_PATH")
	viper.BindEnv("log.es-url", "LOG_ES_URL")
	viper.BindEnv("log.es-index", "LOG_ES_INDEX")
}

func RegistGitStorageFlags(flag *pflag.FlagSet) {
	flag.StringP("git-storage", "s", "", "path to git storage location")
	viper.BindPFlag("git.storage", flag.Lookup("git-storage"))
	viper.BindEnv("git.storage", "GIT_STORAGE_PATH")
}

func RegistGithubTokenFlags(flag *pflag.FlagSet) {
	flag.String("github-token", "", "github token")
	viper.BindPFlag("token.github", flag.Lookup("github-token"))
	viper.BindEnv("token.github", "GITHUB")
}

func RegistRpcFlags(flag *pflag.FlagSet, collector bool, workflow bool) {
	if collector {
		flag.String("rpc-collector", "", "")
		viper.BindPFlag("rpc.collector", flag.Lookup("rpc-collector"))
	}
	if workflow {
		flag.String("rpc-workflow", "", "")
		viper.BindPFlag("rpc.workflow", flag.Lookup("rpc-workflow"))
	}
}

func RegistWebFlags(flag *pflag.FlagSet) {
	flag.String("web-github-oauth-client-id", "", "github oauth client id")
	flag.String("web-github-oauth-client-secret", "", "github oauth client secret")
	flag.String("web-tool-history-dir", "./tool_history", "tool history dir")
	viper.BindPFlag("web.github-oauth-client", flag.Lookup("web-github-oauth-client-id"))
	viper.BindPFlag("web.github-oauth-secret", flag.Lookup("web-github-oauth-client-secret"))
	viper.BindPFlag("web.tool-history-dir", flag.Lookup("web-tool-history-dir"))
}

func RegistWorkflowRunnerFlags(flag *pflag.FlagSet) {
	flag.String("workflow-runner-history-dir", "./workflow_history", "workflow history dir")
	viper.BindPFlag("workflow.history-dir", flag.Lookup("workflow-runner-history-dir"))
}

// include config file, database, log
func RegistCommonFlags(flag *pflag.FlagSet) {
	RegistConfigFileFlags(flag)
	RegistDatabaseFlags(flag)
	RegistLogFlags(flag)
}

// parse flags and set config
func ParseFlags(flag *pflag.FlagSet) {
	// set flag error handling to continue
	rf := reflect.ValueOf(flag).Elem().FieldByName("errorHandling")
	rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
	rf.Set(reflect.ValueOf(pflag.ContinueOnError))
	flag.SortFlags = false

	err := flag.Parse(os.Args[1:])
	if err == pflag.ErrHelp {
		os.Exit(0)
	} else {
		if err != nil {
			logger.Fatalf("Failed to parse flags: %v", err)
		}
	}

	configFile := viper.GetString("config-file")
	if configFile != "" {
		viper.SetConfigFile(configFile)
		viper.ReadInConfig()
	}

	if databaseRegisted {
		storage.InitDefaultDatabaseContext(GetDatabaseConfig())
	}

	if logRegisted {
		logger.Config(GetLogConfig())
	}

}
