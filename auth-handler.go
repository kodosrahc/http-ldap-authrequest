package main

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/go-kit/log/level"
)

const HeaderHttpLdapAuthrequestRequiredGroup = "X-Http-Ldap-Authrequest-RequiredGroup"

func auth_handler(w http.ResponseWriter, r *http.Request) {
	//
	header := r.Header.Get("Authorization")

	if header != "" {
		auth := strings.SplitN(header, " ", 2)

		if len(auth) == 2 && auth[0] == "Basic" {
			decoded, err := base64.StdEncoding.DecodeString(auth[1])
			if err == nil {
				creds := strings.SplitN(string(decoded), ":", 2)

				requiredGroup := r.Header.Get(HeaderHttpLdapAuthrequestRequiredGroup)
				if requiredGroup == "" {
					requiredGroup = config.UserRequiredGroup
				}

				if len(creds) == 2 && validate(creds[0], creds[1], requiredGroup) {
					w.WriteHeader(http.StatusOK)
					return
				}
			} else {
				level.Info(logger).Log("msg", "Error decode basic auth", "err", err)
			}
		}
	}

	w.Header().Set("WWW-Authenticate", config.BasicAuthRealm)
	//w.Header().Set("Cache-Control", "no-cache")

	w.WriteHeader(http.StatusUnauthorized)

}
