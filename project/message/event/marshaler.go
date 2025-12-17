package event

import "github.com/ThreeDotsLabs/watermill/components/cqrs"

var (
	jsonMarshaler = cqrs.JSONMarshaler{
		GenerateName: cqrs.StructName,
	}
)
