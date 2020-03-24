package uploadpack

import (
	"bytes"
	"errors"

	"io"
	"net/http"

	"gitlab.com/gitlab-org/gitlab-shell/internal/gitlabnet"
	"gitlab.com/gitlab-org/gitlab-shell/internal/gitlabnet/accessverifier"
)

type Request struct {
	SecretToken []byte                           `json:"secret_token"`
	Data        accessverifier.CustomPayloadData `json:"data"`
	Output      []byte                           `json:"output"`
}

type Response struct {
	Result  []byte `json:"result"`
	Message string `json:"message"`
}

var readByteSize = 32

func (c *Command) processCustomAction(response *accessverifier.Response) error {
	data := response.Payload.Data
	apiEndpoints := data.ApiEndpoints

	if len(apiEndpoints) == 0 {
		return errors.New("Custom action error: Empty API endpoints")
	}

	return c.processApiEndpoints(response)
}

func (c *Command) processApiEndpoints(response *accessverifier.Response) error {
	client, err := gitlabnet.GetClient(c.Config)

	if err != nil {
		return err
	}

	data := response.Payload.Data
	request := &Request{Data: data}
	request.Data.UserId = response.Who

	for _, endpoint := range data.ApiEndpoints {
		response, err := c.performRequest(client, endpoint, request)

		if err != nil {
			return err
		}

		err = c.displayResult(response.Result)

		if err != nil {
			return err
		}

		// In the context of the git push sequence of events, it's necessary to read
		// stdin in order to capture output to pass onto subsequent commands
		//
		var output bytes.Buffer

		p := make([]byte, readByteSize)

		for {
			n, err := c.ReadWriter.In.Read(p)
			str := string(p[:n])

			if err == io.EOF {
				break
			}

			output.WriteString(str)

			if n < readByteSize {
				break
			}
		}

		request.Output = output.Bytes()
	}

	return nil
}

func (c *Command) performRequest(client *gitlabnet.GitlabClient, endpoint string, request *Request) (*Response, error) {
	response, err := client.DoRequest(http.MethodPost, endpoint, request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	cr := &Response{}
	if err := gitlabnet.ParseJSON(response, cr); err != nil {
		return nil, err
	}

	return cr, nil
}

func (c *Command) displayResult(result []byte) error {
	_, err := io.Copy(c.ReadWriter.Out, bytes.NewReader(result))
	return err
}
