package es

import (
	"encoding/json"
	"fmt"
)

type ILMPhase struct {
	MinAge  string
	Actions map[string]interface{}
}

type ILMPolicy struct {
	Name         string
	Version      int
	ModifiedDate string
	Managed      bool
	Phases       map[string]ILMPhase // keys: "hot", "warm", "cold", "delete"
}

func (c *Client) ListILMPolicies() ([]ILMPolicy, error) {
	data, err := c.Get("/_ilm/policy")
	if err != nil {
		return nil, fmt.Errorf("failed to list ILM policies: %w", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse ILM response: %w", err)
	}

	var policies []ILMPolicy
	for name, val := range raw {
		policyData, ok := val.(map[string]interface{})
		if !ok {
			continue
		}

		p := ILMPolicy{
			Name:         name,
			Version:      jsonInt(policyData["version"]),
			ModifiedDate: JsonStr(policyData["modified_date"]),
			Phases:       make(map[string]ILMPhase),
		}

		if policy, ok := policyData["policy"].(map[string]interface{}); ok {
			if meta, ok := policy["_meta"].(map[string]interface{}); ok {
				if managed, ok := meta["managed"].(bool); ok && managed {
					p.Managed = true
				}
			}
			if phases, ok := policy["phases"].(map[string]interface{}); ok {
				for phaseName, phaseVal := range phases {
					if phaseData, ok := phaseVal.(map[string]interface{}); ok {
						phase := ILMPhase{
							MinAge: JsonStr(phaseData["min_age"]),
						}
						if actions, ok := phaseData["actions"].(map[string]interface{}); ok {
							phase.Actions = actions
						}
						p.Phases[phaseName] = phase
					}
				}
			}
		}

		policies = append(policies, p)
	}

	return policies, nil
}

func (c *Client) CreateILMPolicy(name string, body string) error {
	_, err := c.Put("/_ilm/policy/"+name, body)
	if err != nil {
		return fmt.Errorf("failed to create ILM policy %s: %w", name, err)
	}
	return nil
}

func (c *Client) GetILMPolicy(name string) ([]byte, error) {
	data, err := c.Get("/_ilm/policy/" + name)
	if err != nil {
		return nil, fmt.Errorf("failed to get ILM policy %s: %w", name, err)
	}
	return data, nil
}

func (c *Client) DeleteILMPolicy(name string) error {
	_, err := c.Delete("/_ilm/policy/" + name)
	if err != nil {
		return fmt.Errorf("failed to delete ILM policy %s: %w", name, err)
	}
	return nil
}
