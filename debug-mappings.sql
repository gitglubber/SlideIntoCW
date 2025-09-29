-- Debug script to see all client mappings
SELECT
    slide_client_id,
    slide_client_name,
    connectwise_id,
    connectwise_name,
    created_at
FROM client_mappings
ORDER BY created_at;