UPDATE conversation_messages
SET info = jsonb_set(
    info,
    '{feedback_type}',
    CASE (info->>'feedback_type')::int
        WHEN 1 THEN to_jsonb('内容不准确'::text)
        WHEN 2 THEN to_jsonb('没有帮助'::text)
        WHEN 3 THEN to_jsonb('其他'::text)
        ELSE to_jsonb(''::text)
    END
)
WHERE (info->>'feedback_type') IS NOT NULL;

