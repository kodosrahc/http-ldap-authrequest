package main

type Config struct {
	Listen                    string
	LdapURI                   string
	LdapStartTLS              bool
	LdapTLSInsecureSkipVerify bool
	BindDN                    string
	BindPW                    string
	UserBaseDN                string
	UserFilter                string
	UserRequiredGroup         string
	GroupBaseDN               string
	GroupUserAttr             string
	GroupAttr                 string
	GroupFilter               string
	BasicAuthRealm            string
}
