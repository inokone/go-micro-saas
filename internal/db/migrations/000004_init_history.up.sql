create table microsaas.history_events(
    history_event_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL references microsaas.users(user_id),
    event_type VARCHAR(255) NOT NULL,
    event_data JSONB,
    event_time TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW()
);


