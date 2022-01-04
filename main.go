package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

//VERSION set with -ldflags "-X main.VERSION=$(git describe --long --dirty)"
var VERSION string

var logger log.Logger

var config Config

func main() {
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.Caller(5))

	flag.StringVar(&config.Listen, "listen", "0.0.0.0:8242", "serve http on")
	flag.StringVar(&config.LdapURI, "ldapuri", "", "LDAP URI(s)")
	flag.BoolVar(&config.LdapTLSInsecureSkipVerify, "ldap-tls-insecure-skip-verify", false, "skip ldap TLS verify. Use it only for debug")
	flag.StringVar(&config.BindDN, "auth-bind-dn", "", "auth bind dn")
	fBindPW := flag.String("auth-bind-pw", "", "path to file with bind password")
	flag.StringVar(&config.UserBaseDN, "user-base-dn", "", "search base for user search")
	flag.StringVar(&config.UserFilter, "user-filter", "(cn=%s)", "filter for user search")
	flag.StringVar(&config.UserRequiredGroup, "user-required-group", "", "default required group, if Header X-Http-Ldap-Authrequest-RequiredGroup is not set")
	flag.StringVar(&config.GroupBaseDN, "group-base-dn", "", "search base for group search")
	flag.StringVar(&config.GroupUserAttr, "group-user-attr", "", "identifies user in group membership. If empty, distinguished name is used")
	flag.StringVar(&config.GroupAttr, "group-attr", "cn", "attribute identifying the group. If empty, distinguished name is used")
	flag.StringVar(&config.GroupFilter, "group-filter", "(member=%s)", "filter for group search, the user is member of")
	fRealm := flag.String("realm", "Restricted", "")
	config.BasicAuthRealm = fmt.Sprintf("Basic realm=\"%s\"", *fRealm)

	fVersion := flag.Bool("version", false, "print version and exit")
	fLoglevel := flag.String("loglevel", "warn", "possible values are error, warn, info, debug")
	fTLSCert := flag.String("tls-cert", "", "certificate file")
	fTLSKey := flag.String("tls-key", "", "key file")

	flag.Parse()

	if *fVersion {
		fmt.Println(VERSION)
		os.Exit(0)
	}
	var _allow level.Option
	switch strings.ToLower(*fLoglevel) {
	case "error":
		_allow = level.AllowError()
	case "warn":
		_allow = level.AllowWarn()
	case "info":
		_allow = level.AllowInfo()
	case "debug":
		_allow = level.AllowDebug()
	}
	logger = level.NewFilter(logger, _allow)
	serveTLS := *fTLSCert != "" && *fTLSKey != ""

	level.Info(logger).Log("msg", "Starting http-ldap-authrequest", "version", VERSION, "listen", config.Listen, "tls", serveTLS)

	_bindPW, err := ioutil.ReadFile(*fBindPW)
	config.BindPW = strings.TrimSpace(string(_bindPW))
	if err != nil {
		level.Error(logger).Log("msg", "could not read auth-bind-pw file", "err", err)
		os.Exit(1)
	}

	http.HandleFunc("/", auth_handler)

	if serveTLS {
		err = http.ListenAndServeTLS(config.Listen, *fTLSCert, *fTLSKey, nil)
	} else {
		err = http.ListenAndServe(config.Listen, nil)
	}

	level.Error(logger).Log("msg", "shutting down", "exit code", err)

}
