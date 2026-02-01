-- 支付模块：商品/订单/回调审计/权益事件

-- 商品表
CREATE TABLE IF NOT EXISTS payment_products (
    id BIGSERIAL PRIMARY KEY,
    kind VARCHAR(20) NOT NULL, -- subscription | balance
    name TEXT NOT NULL,
    description_md TEXT NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'inactive', -- active | inactive
    sort_order INT NOT NULL DEFAULT 0,

    currency VARCHAR(10) NOT NULL DEFAULT 'CNY',
    price_cents BIGINT NOT NULL DEFAULT 0,

    -- 订阅商品
    group_id BIGINT DEFAULT NULL REFERENCES groups(id) ON DELETE SET NULL,
    validity_days INT DEFAULT NULL,

    -- 充值商品
    credit_balance DECIMAL(20,8) DEFAULT NULL,
    allow_custom_amount BOOLEAN NOT NULL DEFAULT FALSE,
    min_amount_cents BIGINT DEFAULT NULL,
    max_amount_cents BIGINT DEFAULT NULL,
    suggested_amounts_cents JSONB DEFAULT NULL,
    exchange_rate DECIMAL(20,8) DEFAULT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 订单表
CREATE TABLE IF NOT EXISTS payment_orders (
    id BIGSERIAL PRIMARY KEY,
    order_no TEXT NOT NULL UNIQUE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    kind VARCHAR(20) NOT NULL, -- subscription | balance
    product_id BIGINT DEFAULT NULL REFERENCES payment_products(id) ON DELETE SET NULL,

    status VARCHAR(20) NOT NULL DEFAULT 'created', -- created | paid | fulfilled | cancelled | expired | failed
    provider VARCHAR(20) NOT NULL DEFAULT 'manual', -- epay | tokenpay | manual

    currency VARCHAR(10) NOT NULL DEFAULT 'CNY',
    amount_cents BIGINT NOT NULL,

    client_request_id TEXT DEFAULT NULL,
    provider_trade_no TEXT DEFAULT NULL,
    pay_url TEXT DEFAULT NULL,

    expires_at TIMESTAMPTZ DEFAULT NULL,
    paid_at TIMESTAMPTZ DEFAULT NULL,
    fulfilled_at TIMESTAMPTZ DEFAULT NULL,

    -- 发放快照
    grant_group_id BIGINT DEFAULT NULL REFERENCES groups(id) ON DELETE SET NULL,
    grant_validity_days INT DEFAULT NULL,
    grant_credit_balance DECIMAL(20,8) DEFAULT NULL,

    notes TEXT DEFAULT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(user_id, client_request_id),
    UNIQUE(provider, provider_trade_no)
);

-- 回调审计表（幂等门闩）
CREATE TABLE IF NOT EXISTS payment_notifications (
    id BIGSERIAL PRIMARY KEY,
    provider VARCHAR(20) NOT NULL,
    event_id TEXT NOT NULL,

    order_no TEXT DEFAULT NULL,
    provider_trade_no TEXT DEFAULT NULL,
    amount_cents BIGINT DEFAULT NULL,
    currency VARCHAR(10) DEFAULT NULL,

    verified BOOLEAN NOT NULL DEFAULT FALSE,
    processed BOOLEAN NOT NULL DEFAULT FALSE,
    process_error TEXT DEFAULT NULL,

    raw_body TEXT NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(provider, event_id)
);

-- 权益事件表（区分 redeem/paid/manual）
CREATE TABLE IF NOT EXISTS entitlement_events (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    kind VARCHAR(20) NOT NULL, -- subscription | balance | concurrency
    source VARCHAR(20) NOT NULL, -- redeem | paid | manual

    group_id BIGINT DEFAULT NULL REFERENCES groups(id) ON DELETE SET NULL,
    validity_days INT DEFAULT NULL,
    balance_delta DECIMAL(20,8) DEFAULT NULL,
    concurrency_delta INT DEFAULT NULL,

    order_id BIGINT DEFAULT NULL REFERENCES payment_orders(id) ON DELETE SET NULL,
    redeem_code_id BIGINT DEFAULT NULL REFERENCES redeem_codes(id) ON DELETE SET NULL,
    actor_user_id BIGINT DEFAULT NULL REFERENCES users(id) ON DELETE SET NULL,

    note TEXT DEFAULT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_payment_products_status_sort ON payment_products(status, sort_order);
CREATE INDEX IF NOT EXISTS idx_payment_products_kind_status ON payment_products(kind, status);
CREATE INDEX IF NOT EXISTS idx_payment_products_group_id ON payment_products(group_id);

CREATE INDEX IF NOT EXISTS idx_payment_orders_user_created_at ON payment_orders(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_payment_orders_status_created_at ON payment_orders(status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_payment_orders_product_id ON payment_orders(product_id);
CREATE INDEX IF NOT EXISTS idx_payment_orders_provider_trade_no ON payment_orders(provider, provider_trade_no);

CREATE INDEX IF NOT EXISTS idx_payment_notifications_order_no ON payment_notifications(order_no);
CREATE INDEX IF NOT EXISTS idx_payment_notifications_provider_trade_no ON payment_notifications(provider, provider_trade_no);
CREATE INDEX IF NOT EXISTS idx_payment_notifications_received_at ON payment_notifications(received_at);

CREATE INDEX IF NOT EXISTS idx_entitlement_events_user_id ON entitlement_events(user_id);
CREATE INDEX IF NOT EXISTS idx_entitlement_events_order_id ON entitlement_events(order_id);
CREATE INDEX IF NOT EXISTS idx_entitlement_events_redeem_code_id ON entitlement_events(redeem_code_id);
CREATE INDEX IF NOT EXISTS idx_entitlement_events_created_at ON entitlement_events(created_at);

COMMENT ON TABLE payment_products IS '支付商品';
COMMENT ON TABLE payment_orders IS '支付订单';
COMMENT ON TABLE payment_notifications IS '支付回调审计与幂等门闩';
COMMENT ON TABLE entitlement_events IS '权益发放事件（来源审计）';
