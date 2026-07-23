UPDATE conversation_messages
SET info = jsonb_set(
    info,
    '{feedback_type}',
    CASE (info->>'feedback_type')
        WHEN '内容不准确' THEN '1'::jsonb
        WHEN '没有帮助'   THEN '2'::jsonb
        WHEN '其他'       THEN '3'::jsonb
        WHEN '' THEN '0'::jsonb
        ELSE (info->'feedback_type') 
    END
)
WHERE (info->>'feedback_type') IS NOT NULL;