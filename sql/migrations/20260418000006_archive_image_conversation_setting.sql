-- +goose Up
INSERT INTO `system_settings` (`k`, `v`, `description`) VALUES
  ('gateway.archive_image_conversation', 'false', '图片结果可下载后自动把 chatgpt.com 对话设为归档,避免污染官网会话列表')
ON DUPLICATE KEY UPDATE `description` = VALUES(`description`);

-- +goose Down
DELETE FROM `system_settings`
WHERE `k` = 'gateway.archive_image_conversation';
