package es

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

type Task struct {
	ID               string
	NodeID           string
	NodeName         string
	Action           string
	Type             string
	Description      string
	RunningTimeNanos int64
	StartTimeMillis  int64
	Cancellable      bool
	ParentTaskID     string
}

func FormatTaskDuration(nanos int64) string {
	if nanos < 1_000 {
		return fmt.Sprintf("%dns", nanos)
	}
	if nanos < 1_000_000 {
		return fmt.Sprintf("%dµs", nanos/1_000)
	}
	if nanos < 1_000_000_000 {
		return fmt.Sprintf("%.1fms", float64(nanos)/1_000_000)
	}
	if nanos < 60_000_000_000 {
		return fmt.Sprintf("%.1fs", float64(nanos)/1_000_000_000)
	}
	mins := nanos / 60_000_000_000
	secs := (nanos % 60_000_000_000) / 1_000_000_000
	return fmt.Sprintf("%dm%02ds", mins, secs)
}

func FormatTaskStartTime(millis int64) string {
	if millis == 0 {
		return ""
	}
	return time.Unix(millis/1000, (millis%1000)*1_000_000).Local().Format("2006-01-02 15:04:05 MST")
}

func (c *Client) ListTasks() ([]Task, int, error) {
	data, err := c.Get("/_tasks?detailed=true")
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list tasks: %w", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, 0, fmt.Errorf("failed to parse tasks response: %w", err)
	}

	var tasks []Task
	nodes, _ := raw["nodes"].(map[string]interface{})
	for nodeID, nodeData := range nodes {
		nodeMap, _ := nodeData.(map[string]interface{})
		nodeName := JsonStr(nodeMap["name"])
		tasksMap, _ := nodeMap["tasks"].(map[string]interface{})
		for taskID, taskData := range tasksMap {
			taskMap, _ := taskData.(map[string]interface{})
			cancellable, _ := taskMap["cancellable"].(bool)
			tasks = append(tasks, Task{
				ID:               taskID,
				NodeID:           nodeID,
				NodeName:         nodeName,
				Action:           JsonStr(taskMap["action"]),
				Type:             JsonStr(taskMap["type"]),
				Description:      JsonStr(taskMap["description"]),
				RunningTimeNanos: jsonInt64(taskMap["running_time_in_nanos"]),
				StartTimeMillis:  jsonInt64(taskMap["start_time_in_millis"]),
				Cancellable:      cancellable,
				ParentTaskID:     JsonStr(taskMap["parent_task_id"]),
			})
		}
	}

	total := len(tasks)
	sort.SliceStable(tasks, func(i, j int) bool {
		return tasks[i].RunningTimeNanos > tasks[j].RunningTimeNanos
	})
	return tasks, total, nil
}

func (c *Client) CancelTask(taskID string) error {
	_, err := c.Post("/_tasks/"+taskID+"/_cancel", "")
	if err != nil {
		return fmt.Errorf("failed to cancel task: %w", err)
	}
	return nil
}

func (c *Client) GetTaskDetail(taskID string) ([]byte, error) {
	data, err := c.Get("/_tasks/" + taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task detail: %w", err)
	}
	return transformTaskJSON(data)
}

func transformTaskJSON(data []byte) ([]byte, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return data, nil
	}

	if task, ok := raw["task"].(map[string]interface{}); ok {
		if nanos := jsonInt64(task["running_time_in_nanos"]); nanos > 0 {
			task["running_time"] = FormatTaskDuration(nanos)
			delete(task, "running_time_in_nanos")
		}
		if millis := jsonInt64(task["start_time_in_millis"]); millis > 0 {
			task["start_time"] = FormatTaskStartTime(millis)
			delete(task, "start_time_in_millis")
		}
		raw["task"] = task
	}

	result, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return data, nil
	}
	return result, nil
}
