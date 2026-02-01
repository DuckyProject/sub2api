package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// PaymentNotification holds the schema definition for the PaymentNotification entity.
//
// 删除策略：硬删除
type PaymentNotification struct {
	ent.Schema
}

func (PaymentNotification) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "payment_notifications"},
	}
}

func (PaymentNotification) Fields() []ent.Field {
	return []ent.Field{
		field.String("provider").
			MaxLen(20),
		field.String("event_id").
			NotEmpty(),

		field.String("order_no").
			Optional().
			Nillable(),
		field.String("provider_trade_no").
			Optional().
			Nillable(),
		field.Int64("amount_cents").
			Optional().
			Nillable(),
		field.String("currency").
			MaxLen(10).
			Optional().
			Nillable(),

		field.Bool("verified").
			Default(false),
		field.Bool("processed").
			Default(false),
		field.String("process_error").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),

		field.String("raw_body").
			NotEmpty().
			SchemaType(map[string]string{dialect.Postgres: "text"}),

		field.Time("received_at").
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (PaymentNotification) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("provider", "event_id").
			Unique(),
		index.Fields("order_no"),
		index.Fields("provider_trade_no"),
		index.Fields("received_at"),
	}
}
