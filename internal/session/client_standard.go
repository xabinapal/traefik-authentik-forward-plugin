package session

type StandardClient struct {
}

func NewStandardClient() *StandardClient {
	return &StandardClient{}
}

func (c *StandardClient) Get(session string) *Session {
	return nil
}

func (c *StandardClient) Set(session string, meta *Session) {
}

func (c *StandardClient) Delete(session string) {
}
