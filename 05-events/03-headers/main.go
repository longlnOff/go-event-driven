package main

import (
	"encoding/json"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
)

type MessageHeader struct {
	ID             string `json:"id"`
	EventName      string `json:"event_name"`
	OccurredAt 	   string `json:"occurred_at"`

}

func NewMessageHeader(eventName string) MessageHeader {
	return MessageHeader{
		ID:          uuid.NewString(), 
		EventName:   eventName, 
		OccurredAt: time.Now().Format(time.RFC3339),
	}
}


type ProductOutOfStock struct {
	Header  MessageHeader `json:"header"`
	ProductID string `json:"product_id"`
}

type ProductBackInStock struct {
	Header  MessageHeader `json:"header"`
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type Publisher struct {
	pub message.Publisher
}

func NewPublisher(pub message.Publisher) Publisher {
	return Publisher{
		pub: pub,
	}
}

func (p Publisher) PublishProductOutOfStock(productID string) error {
	event := ProductOutOfStock{
		Header: NewMessageHeader("ProductOutOfStock"),
		ProductID: productID,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)

	return p.pub.Publish("product-updates", msg)
}

func (p Publisher) PublishProductBackInStock(productID string, quantity int) error {
	event := ProductBackInStock{
		Header: NewMessageHeader("ProductBackInStock"),
		ProductID: productID,
		Quantity:  quantity,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)

	return p.pub.Publish("product-updates", msg)
}
