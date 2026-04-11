package es

import (
	"encoding/json"
	"fmt"
)

type Shard struct {
	Index  string
	ShardN string
	PriRep string
	State  string
	Docs   string
	Store  string
	IP     string
	Node   string
}

func (c *Client) ListShards() ([]Shard, error) {
	data, err := c.Get("/_cat/shards?format=json&h=index,shard,prirep,state,docs,store,ip,node")
	if err != nil {
		return nil, fmt.Errorf("failed to list shards: %w", err)
	}

	var raw []map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse shards response: %w", err)
	}

	shards := make([]Shard, len(raw))
	for i, r := range raw {
		shards[i] = Shard{
			Index:  JsonStr(r["index"]),
			ShardN: JsonStr(r["shard"]),
			PriRep: JsonStr(r["prirep"]),
			State:  JsonStr(r["state"]),
			Docs:   JsonStr(r["docs"]),
			Store:  JsonStr(r["store"]),
			IP:     JsonStr(r["ip"]),
			Node:   JsonStr(r["node"]),
		}
	}

	return shards, nil
}
