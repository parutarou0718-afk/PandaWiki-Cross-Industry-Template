-- Create settings table
CREATE TABLE IF NOT EXISTS system_settings (
    id SERIAL PRIMARY KEY,
    key TEXT NOT NULL,
    value JSONB NOT NULL,
    description TEXT,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_uniq_system_settings_key ON system_settings(key);

-- Insert model_setting_mode setting
-- If there are existing knowledge bases, set mode to 'manual', otherwise set to 'auto'
INSERT INTO system_settings (key, value, description)
SELECT 
    'model_setting_mode',
    jsonb_build_object(
        'mode', CASE 
            WHEN EXISTS (SELECT 1 FROM knowledge_bases LIMIT 1) THEN 'manual'
            ELSE 'auto'
        END,
        'auto_mode_api_key', '',
        'chat_model', '',
        'is_manual_embedding_updated', false
    ),
    'Model setting mode configuration'
WHERE NOT EXISTS (
    SELECT 1 FROM system_settings WHERE key = 'model_setting_mode'
);