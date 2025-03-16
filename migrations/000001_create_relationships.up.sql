CREATE TABLE IF NOT EXISTS relationships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL, -- Supabase Auth UUID
    other_user_id UUID NOT NULL, -- Supabase Auth UUID
    relationship_status VARCHAR(20) CHECK (relationship_status IN 
        ('none', 'user', 'friend', 'outgoing', 'incoming', 'blocked', 'blocked_other')),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE (user_id, other_user_id)
);