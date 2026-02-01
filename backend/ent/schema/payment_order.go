package schema

import (
	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// PaymentOrder holds the schema definition for the PaymentOrder entity.
//
// 删除策略：硬删除
type PaymentOrder struct {
	ent.Schema
}

func (PaymentOrder) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "payment_orders"},
	}
}

func (PaymentOrder) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
	}
}

func (PaymentOrder) Fields() []ent.Field {
	return []ent.Field{
		field.String("order_no").
			NotEmpty().
			Unique(),
		field.Int64("user_id"),
		field.String("kind").
			MaxLen(20).
			Comment("subscription | balance"),
		field.Int64("product_id").
			Optional().
			Nillable(),

		field.String("status").
			MaxLen(20).
			Default("created"),
		field.String("provider").
			MaxLen(20).
			Default("manual"),

		field.String("currency").
			MaxLen(10).
			Default("CNY"),
		field.Int64("amount_cents"),

		field.String("client_request_id").
			Optional().
			Nillable(),
		field.String("provider_trade_no").
			Optional().
			Nillable(),
		field.String("pay_url").
			Optional().
			Nillable(),

		field.Time("expires_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("paid_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("fulfilled_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),

		field.Int64("grant_group_id").
			Optional().
			Nillable(),
		field.Int("grant_validity_days").
			Optional().
			Nillable(),
		field.Float("grant_credit_balance").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}),

		field.String("notes").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
	}
}

func (PaymentOrder) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("payment_orders").
			Field("user_id").
			Unique().
			Required(),
		edge.From("product", PaymentProduct.Type).
			Ref("orders").
			Field("product_id").
			Unique(),
		edge.From("grant_group", Group.Type).
			Ref("payment_orders_grant_group").
			Field("grant_group_id").
			Unique(),
		edge.To("entitlement_events", EntitlementEvent.Type),
	}
}

func (PaymentOrder) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "created_at"),
		index.Fields("status", "created_at"),
		index.Fields("product_id"),
		index.Fields("provider", "provider_trade_no").
			Unique(),
		index.Fields("user_id", "client_request_id").
			Unique(),
	}
}
