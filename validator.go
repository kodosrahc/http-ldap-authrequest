package main

import (
	"crypto/tls"
	"fmt"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/go-ldap/ldap/v3"
)

func validate(user, pass, group string) (userValid bool) {
	logger := log.With(logger, "user", user, "group", group)
	level.Debug(logger).Log("msg", "will validate")
	defer func() {
		if userValid {
			level.Info(logger).Log("msg", "user is authenticated")
		} else {
			level.Info(logger).Log("msg", "user is denied")
		}
	}()

	var tlsConfig tls.Config

	if config.LdapTLSInsecureSkipVerify {
		tlsConfig = tls.Config{InsecureSkipVerify: true}
	}

	l, err := ldap.DialURL(config.LdapURI, ldap.DialWithTLSConfig(&tlsConfig))
	if err != nil {
		level.Error(logger).Log("msg", "could not dial LDAP", "err", err)
		return false
	}
	defer l.Close()

	if config.LdapStartTLS {
		err := l.StartTLS(&tlsConfig)
		if err != nil {
			level.Error(logger).Log("msg", "could not StartTLS", "err", err)
			return false
		}
	}

	// ldap bind
	err = l.Bind(config.BindDN, config.BindPW)
	if err != nil {
		level.Error(logger).Log("msg", "could not auth bind", "err", err)
		return false
	}

	//user lookup
	filter := fmt.Sprintf(config.UserFilter, ldap.EscapeFilter(user))
	lrequest := ldap.NewSearchRequest(config.UserBaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		filter,
		[]string{"1.1"},
		nil)
	lresult, err := l.Search(lrequest)
	if err != nil {
		level.Error(logger).Log("msg", "error while searching user", "err", err)
		return false
	}

	if len(lresult.Entries) < 1 {
		level.Debug(logger).Log("msg", "user not found")
		return false
	}

	if len(lresult.Entries) > 1 {
		level.Error(logger).Log("msg", "ambiguous user search result", "#entries", len(lresult.Entries))
		return false
	}

	userDN := lresult.Entries[0].DN
	level.Debug(logger).Log("msg", "user found", "dn", userDN)

	//lookup if the user is the member of the specific group
	if group != "" {
		filter = fmt.Sprintf(config.GroupFilter, ldap.EscapeFilter(userDN))
		lrequest = ldap.NewSearchRequest(config.GroupBaseDN,
			ldap.ScopeWholeSubtree,
			ldap.NeverDerefAliases,
			0,
			0,
			false,
			filter,
			[]string{config.GroupAttr},
			nil)
		lresult, err = l.Search(lrequest)
		if err != nil {
			level.Error(logger).Log("msg", "error while searching group", "err", err)
			return false
		}

		var groupMembershipConfirmed bool = false
		for _, e := range lresult.Entries {
			if e.GetAttributeValue(config.GroupAttr) == group {
				groupMembershipConfirmed = true
				break
			}
		}
		if !groupMembershipConfirmed {
			level.Debug(logger).Log("msg", "no required group found for the user")
			return false
		}

		level.Debug(logger).Log("msg", "found the required group for the user")
	}

	//user bind
	err = l.Bind(userDN, pass)
	if err != nil {
		level.Debug(logger).Log("msg", "user bind failed", "err", err)
		return false
	}

	return true
}
