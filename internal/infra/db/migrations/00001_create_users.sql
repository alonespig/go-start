-- +goose Up
CREATE TABLE IF NOT EXISTS `users` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `name` VARCHAR(64) NOT NULL,
    `email` VARCHAR(128) NOT NULL,
    `age` INT NOT NULL DEFAULT 0,
    `created_at` DATETIME(3) NULL,
    `updated_at` DATETIME(3) NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_users_email` (`email`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;

-- +goose Down
DROP TABLE IF EXISTS `users`;
