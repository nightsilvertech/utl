package vault

import (
	"fmt"
	"github.com/hashicorp/vault/api"
	"strings"
)

type VLT interface {
	Get(string) string
}

type vlt struct {
	path    string
	client  *api.Logical
	results map[string]map[string]string
}

func (p *vlt) Get(v string) string {
	// <path>/data/<path-secret>:key
	split := strings.Split(v, ":")
	if len(split) == 1 {
		return ""
	}

	pathSecret := split[0]
	key := split[1]

	res, ok := p.results[pathSecret]
	if ok {
		val, ok := res[key]
		if !ok {
			return ""
		}

		return val
	}

	secret, err := p.client.Read(fmt.Sprintf("%s/data/%s", p.path, pathSecret))
	if err != nil {
		return ""
	}

	if secret == nil {
		return ""
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return ""
	}

	secrets := make(map[string]string)

	for k, v := range data {
		val, ok := v.(string)
		if !ok {
			return ""
		}

		secrets[k] = val
	}

	val, ok := secrets[key]
	if !ok {
		return ""
	}

	p.results[pathSecret] = secrets

	return val
}

func NewVLT(token, addr, path string) (VLT, error) {
	config := &api.Config{Address: addr}
	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("new client: %w", err)
	}
	client.SetToken(token)
	return &vlt{
		path:    path,
		client:  client.Logical(),
		results: make(map[string]map[string]string),
	}, nil
}

