package cri

import (
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/credentialprovider"
	credentialprovidersecrets "k8s.io/kubernetes/pkg/credentialprovider/secrets"

	"knative.dev/cache-imagewarm/pkg/warmer/utils"
)

var (
	keyring = credentialprovider.NewDockerKeyring()
)

func ConvertToRegistryAuths(pullSecret v1.Secret, repo string) (infos []utils.AuthInfo, err error) {
	keyring, err := credentialprovidersecrets.MakeDockerKeyring([]v1.Secret{pullSecret}, keyring)
	if err != nil {
		return nil, err
	}
	creds, withCredentials := keyring.Lookup(repo)
	if !withCredentials {
		return nil, nil
	}
	for _, c := range creds {
		infos = append(infos, utils.AuthInfo{
			Username: c.Username,
			Password: c.Password,
		})
	}
	return infos, nil
}

// ParseRepositoryTag gets a repos name and returns the right reposName + tag|digest
// The tag can be confusing because of a port in a repository name.
//     Ex: localhost.localdomain:5000/samalba/hipache:latest
//     Digest ex: localhost:5000/foo/bar@sha256:bc8813ea7b3603864987522f02a76101c17ad122e1c46d790efc0fca78ca7bfb

func ParseRepositoryTag(repos string) (string, string) {
	n := strings.Index(repos, "@")
	if n >= 0 {
		parts := strings.Split(repos, "@")
		return parts[0], parts[1]
	}
	n = strings.LastIndex(repos, ":")
	if n < 0 {
		return repos, ""
	}
	if tag := repos[n+1:]; !strings.Contains(tag, "/") {
		return repos[:n], tag
	}
	return repos, ""
}
