package hw10programoptimization

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
)

type User struct {
	Email string
}

type DomainStat map[string]int

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	domainStat := make(DomainStat)
	domainSuffix := "." + strings.ToLower(domain)
	decoder := json.NewDecoder(r)

	for {
		var user User
		err := decoder.Decode(&user)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}

		atIndex := strings.LastIndex(user.Email, "@")
		if atIndex == -1 {
			continue
		}

		emailDomain := strings.ToLower(user.Email[atIndex+1:])
		if !strings.HasSuffix(emailDomain, domainSuffix) {
			continue
		}

		domainStat[emailDomain]++
	}

	return domainStat, nil
}
