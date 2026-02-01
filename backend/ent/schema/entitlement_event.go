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

// EntitlementEvent holds the schema definition for the EntitlementEvent entity.
//
// 删除策略：硬删除
type EntitlementEvent struct {
	ent.Schema
}

func (EntitlementEvent) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "entitlement_events"},
	}
}

func (EntitlementEvent) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
		mixins.SoftDeleteMixin{},
	}
}

func (EntitlementEvent) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id"),
		field.String("kind").
			MaxLen(20).
			Comment("subscription | balance | concurrency"),
		field.String("source").
			MaxLen(20).
			Comment("redeem | paid | manual"),

		field.Int64("group_id").
			Optional().
			Nillable(),
		field.Int("validity_days").
			Optional().
			Nillable(),
		field.Float("balance_delta").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}),
		field.Int("concurrency_delta").
			Optional().
			Nillable(),

		field.Int64("order_id").
			Optional().
			Nillable(),
		field.Int64("redeem_code_id").
			Optional().
			Nillable(),
		field.Int64("actor_user_id").
			Optional().
			Nillable(),

		field.String("note").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
	}
}

func (EntitlementEvent) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("entitlement_events").
			Field("user_id").
			Unique().
			Required(),
		edge.From("group", Group.Type).
			Ref("entitlement_events").
			Field("group_id").
			Unique(),
		edge.From("order", PaymentOrder.Type).
			Ref("entitlement_events").
			Field("order_id").
			Unique(),
		edge.From("redeem_code", RedeemCode.Type).
			Ref("entitlement_events").
			Field("redeem_code_id").
			Unique(),
		edge.From("actor_user", User.Type).
			Ref("entitlement_events_actor_user").
			Field("actor_user_id").
			Unique(),
	}
}

func (EntitlementEvent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("order_id"),
		index.Fields("redeem_code_id"),
		index.Fields("created_at"),
	}
}
