package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/buildkite/agent/v3/internal/pipeline"
)

// Job represents a Buildkite Agent API Job
type Job struct {
	ID                 string               `json:"id,omitempty"`
	Endpoint           string               `json:"endpoint"`
	State              string               `json:"state,omitempty"`
	Env                map[string]string    `json:"env,omitempty"`
	Step               pipeline.CommandStep `json:"step,omitempty"`
	ChunksMaxSizeBytes uint64               `json:"chunks_max_size_bytes,omitempty"`
	LogMaxSizeBytes    uint64               `json:"log_max_size_bytes,omitempty"`
	Token              string               `json:"token,omitempty"`
	ExitStatus         string               `json:"exit_status,omitempty"`
	Signal             string               `json:"signal,omitempty"`
	SignalReason       string               `json:"signal_reason,omitempty"`
	StartedAt          string               `json:"started_at,omitempty"`
	FinishedAt         string               `json:"finished_at,omitempty"`
	RunnableAt         string               `json:"runnable_at,omitempty"`
	ChunksFailedCount  int                  `json:"chunks_failed_count,omitempty"`
}

func (j *Job) ValuesForFields(fields []string) (map[string]string, error) {
	o := make(map[string]string, len(fields))
	for _, f := range fields {
		switch f {
		case "command":
			o[f] = j.Env["BUILDKITE_COMMAND"]

		case "plugins":
			if j.Env["BUILDKITE_PLUGINS"] == "" {
				o[f] = ""
				continue
			}
			// Plugins needs to be normalised, because key order in each plugin
			// config is frequently varied by the backend.
			// The reliable way to make it consistent is an unmarshal-remarshal
			// round-trip.
			var ps pipeline.Plugins
			if err := json.Unmarshal([]byte(j.Env["BUILDKITE_PLUGINS"]), &ps); err != nil {
				return nil, fmt.Errorf("unmarshaling BUIDLKITE_PLUGINS: %w", err)
			}
			normalised, err := json.Marshal(ps)
			if err != nil {
				return nil, fmt.Errorf("re-marshaling BUIDLKITE_PLUGINS: %w", err)
			}
			o[f] = string(normalised)

		default:
			if e, has := strings.CutPrefix(f, pipeline.EnvNamespacePrefix); has {
				o[f] = j.Env[e]
				break
			}

			return nil, fmt.Errorf("unknown or unsupported field on Job struct for signing/verification: %q", f)
		}
	}

	return o, nil
}

type JobState struct {
	State string `json:"state,omitempty"`
}

type jobStartRequest struct {
	StartedAt string `json:"started_at,omitempty"`
}

type jobFinishRequest struct {
	ExitStatus        string `json:"exit_status,omitempty"`
	Signal            string `json:"signal,omitempty"`
	SignalReason      string `json:"signal_reason,omitempty"`
	FinishedAt        string `json:"finished_at,omitempty"`
	ChunksFailedCount int    `json:"chunks_failed_count"`
}

// GetJobState returns the state of a given job
func (c *Client) GetJobState(ctx context.Context, id string) (*JobState, *Response, error) {
	u := fmt.Sprintf("jobs/%s", id)

	req, err := c.newRequest(ctx, "GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	s := new(JobState)
	resp, err := c.doRequest(req, s)
	if err != nil {
		return nil, resp, err
	}

	return s, resp, err
}

// Acquires a job using its ID
func (c *Client) AcquireJob(ctx context.Context, id string, headers ...Header) (*Job, *Response, error) {
	u := fmt.Sprintf("jobs/%s/acquire", id)

	req, err := c.newRequest(ctx, "PUT", u, nil, headers...)
	if err != nil {
		return nil, nil, err
	}

	j := new(Job)
	resp, err := c.doRequest(req, j)
	if err != nil {
		return nil, resp, err
	}

	return j, resp, err
}

// AcceptJob accepts the passed in job. Returns the job with its finalized set of
// environment variables (when a job is accepted, the agents environment is
// applied to the job)
func (c *Client) AcceptJob(ctx context.Context, job *Job) (*Job, *Response, error) {
	u := fmt.Sprintf("jobs/%s/accept", job.ID)

	req, err := c.newRequest(ctx, "PUT", u, nil)
	if err != nil {
		return nil, nil, err
	}

	j := new(Job)
	resp, err := c.doRequest(req, j)
	if err != nil {
		return nil, resp, err
	}

	return j, resp, err
}

// StartJob starts the passed in job
func (c *Client) StartJob(ctx context.Context, job *Job) (*Response, error) {
	u := fmt.Sprintf("jobs/%s/start", job.ID)

	req, err := c.newRequest(ctx, "PUT", u, &jobStartRequest{
		StartedAt: job.StartedAt,
	})
	if err != nil {
		return nil, err
	}

	return c.doRequest(req, nil)
}

// FinishJob finishes the passed in job
func (c *Client) FinishJob(ctx context.Context, job *Job) (*Response, error) {
	u := fmt.Sprintf("jobs/%s/finish", job.ID)

	req, err := c.newRequest(ctx, "PUT", u, &jobFinishRequest{
		FinishedAt:        job.FinishedAt,
		ExitStatus:        job.ExitStatus,
		Signal:            job.Signal,
		SignalReason:      job.SignalReason,
		ChunksFailedCount: job.ChunksFailedCount,
	})
	if err != nil {
		return nil, err
	}

	return c.doRequest(req, nil)
}
