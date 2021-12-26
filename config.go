package main

type Config struct {
	Listen             *string
	LDAPServer         *string
	BindDN             *string
	BindPW             *string
	UserBaseDN         *string
	UserFilter         *string
	UserRequiredGroups *string
	GroupBaseDN        *string
	GroupAttr          *string
	GroupFilter        *string
}
