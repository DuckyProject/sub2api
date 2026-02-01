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

// PaymentProduct holds the schema definition for the PaymentProduct entity.
//
// 删除策略：硬删除
type PaymentProduct struct {
	ent.Schema
}

func (PaymentProduct) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "payment_products"},
	}
}

func (PaymentProduct) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
	}
}

func (PaymentProduct) Fields() []ent.Field {
	return []ent.Field{
		field.String("kind").
			MaxLen(20).
			Comment("subscription | balance"),
		field.String("name").
			NotEmpty(),
		field.String("description_md").
			Default(""),
		field.String("status").
			MaxLen(20).
			Default("inactive").
			Comment("active | inactive"),
		field.Int("sort_order").
			Default(0),

		field.String("currency").
			MaxLen(10).
			Default("CNY"),
		field.Int64("price_cents").
			Default(0),

		field.Int64("group_id").
			Optional().
			Nillable(),
		field.Int("validity_days").
			Optional().
			Nillable(),

		field.Float("credit_balance").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}),
		field.Bool("allow_custom_amount").
			Default(false),
		field.Int64("min_amount_cents").
			Optional().
			Nillable(),
		field.Int64("max_amount_cents").
			Optional().
			Nillable(),
		field.JSON("suggested_amounts_cents", []int64{}).
			Optional().
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.Float("exchange_rate").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}),
	}
}

func (PaymentProduct) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("group", Group.Type).
			Ref("payment_products").
			Field("group_id").
			Unique(),
		edge.To("orders", PaymentOrder.Type),
	}
}

func (PaymentProduct) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status", "sort_order"),
		index.Fields("kind", "status"),
		index.Fields("group_id"),
	}
}
