package core

import "errors"

var ErrDuplicateCompanyDomain = errors.New("company domain already exists")
var ErrActivePrimaryDomainExists = errors.New("company already has an active primary domain")
var ErrCompanyDomainDoesNotBelong = errors.New("company domain does not belong to company")
