package blocks

// New creates a new QueryResponse with the given blocks. This is the top level block.
// and should be used to return all subsections unless you wish to create the struct manually.
func New(blocks ...Block) QueryResponse {
	return QueryResponse{
		CadenceResponseType: "formattedData",
		Format:              "blocks",
		Blocks:              blocks,
	}
}

// Creates a markdown block with the given text.
func NewMarkdownSection(markdownText string) Block {
	return Block{
		Type:   "section",
		Format: "text/markdown",
		ComponentOptions: &ComponentOptions{
			Text: markdownText,
		},
	}
}

// NewDivider creates a divider in the UI.
func NewDivider() Block {
	return Block{
		Type: "divider",
	}
}

// Creates a set of actions for signalling the workflow.
func NewSignalActions(elements ...Element) Block {
	return Block{
		Type:     "actions",
		Elements: elements,
	}
}

// NewSignalButton creates a button that will signal the workflow with the given signal name and value.
// the signal value can be nil if no value is needed.
func NewSignalButton(text string, signalName string, signalValue interface{}) Element {
	return Element{
		Type: "button",
		ComponentOptions: &ComponentOptions{
			Type: "plain_text",
			Text: text,
		},
		Action: &Action{
			Type:        "signal",
			SignalName:  signalName,
			SignalValue: signalValue,
		},
	}
}

// NewSignalButtonWithExternalWorkflow creates a button that will signal the workflow with the given signal name and value,
// and will also start the external workflow with the given workflow ID and run ID.
//
// The RunID should be optional and the latest workflow will be selected i it's not provided
func NewSignalButtonWithExternalWorkflow(text string, signalName string, signalValue interface{}, workflowID string, runID string) Element {
	return Element{
		Type: "button",
		ComponentOptions: &ComponentOptions{
			Type: "plain_text",
			Text: text,
		},
		Action: &Action{
			Type:        "signal",
			SignalName:  signalName,
			SignalValue: signalValue,
			WorkflowID:  workflowID,
			RunID:       runID,
		},
	}
}

// Query response is the overall wrapper struct that will be returned to the client.
// There's nothing special about this Go code specifically, the actual UI
// only cares about the JSON structure returned, but this is a helpful wrapper
type QueryResponse struct {
	CadenceResponseType string  `json:"cadenceResponseType"`
	Format              string  `json:"format"`
	Blocks              []Block `json:"blocks"`
}

// A section in the workflow Query response that will be rendered
type Block struct {
	Type             string            `json:"type"`
	Format           string            `json:"format,omitempty"`
	ComponentOptions *ComponentOptions `json:"componentOptions,omitempty"`
	Elements         []Element         `json:"elements,omitempty"`
}

type Element struct {
	Type             string            `json:"type"`
	ComponentOptions *ComponentOptions `json:"componentOptions,omitempty"`
	Action           *Action           `json:"action,omitempty"`
}

type ComponentOptions struct {
	Type string `json:"type,omitempty"`
	Text string `json:"text,omitempty"`
}

// Action signifying something such as a button
type Action struct {
	Type        string      `json:"type"`
	SignalName  string      `json:"signal_name,omitempty"`
	SignalValue interface{} `json:"signal_value,omitempty"`
	WorkflowID  string      `json:"workflow_id,omitempty"`
	RunID       string      `json:"run_id,omitempty"`
}
