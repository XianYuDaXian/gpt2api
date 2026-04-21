-- +goose Up
INSERT INTO `system_settings` (`k`, `v`, `description`) VALUES
  ('gateway.delete_rejected_image_conversation', 'false', '图片请求被上游明确拒绝时自动删除 chatgpt.com 会话,避免污染官网会话列表')
ON DUPLICATE KEY UPDATE `description` = VALUES(`description`);

-- +goose Down
DELETE FROM `system_settings`
WHERE `k` = 'gateway.delete_rejected_image_conversation';
