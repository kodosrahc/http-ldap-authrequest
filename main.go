package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

//VERSION set with -ldflags "-X main.VERSION=$(git describe --long --dirty)"
var VERSION string

var logger log.Logger

var config Config

func main() {
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = level.NewFilter(logger, level.AllowInfo())
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
	level.Info(logger).Log("msg", "http-ldap-authrequest", "version", VERSION)

	config.Listen = flag.String("listen", "0.0.0.0:8242", "serve http on")
	config.LDAPServer = flag.String("ldap", "", "LDAP URI")
	config.BindDN = flag.String("auth-bind-dn", "", "auth bind dn")
	config.BindPW = flag.String("auth-bind-pw", "", "path to file with bind password")
	config.UserBaseDN = flag.String("user-base-dn", "", "search base for user search")
	config.UserFilter = flag.String("user-filter", "(cn={0})", "filter for user search")
	config.UserRequiredGroups = flag.String("user-required-groups", "", "required groups")
	config.GroupBaseDN = flag.String("group-base-dn", "", "search base for group search")
	config.GroupAttr = flag.String("group-attr", "cn", "")
	config.GroupFilter = flag.String("group-filter", "(member={0})", "filter for group search")

	fVersion := flag.Bool("version", false, "print version and exit")

	flag.Parse()

	if *fVersion {
		fmt.Println(VERSION)
		os.Exit(0)
	}

	level.Info(logger).Log("msg", "starting plain http ListenAndServe", "address", config.Listen)

	http.HandleFunc("/", auth_handler)

	err := http.ListenAndServe(*config.Listen, nil)

	level.Error(logger).Log("msg", "shutting down", "exit code", err)

}
