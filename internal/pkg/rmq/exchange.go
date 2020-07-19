package rmq

type exchangeType string

const (
	// DIRECT exchange type
	DIRECT exchangeType = "direct"
	// TOPIC exchange type
	TOPIC exchangeType = "topic"
	// HEADERS exchange type
	HEADERS exchangeType = "headers"
	// FANOUT exchange type
	FANOUT exchangeType = "fanout"
)

// Exchange for rabbitmq configuration
type Exchange struct {
	ExcName string
	ExcType ExchangeType
}

// ExchangeType interface
type ExchangeType interface {
	Type() exchangeType
}

func (b exchangeType) Type() exchangeType {
	return b
}
