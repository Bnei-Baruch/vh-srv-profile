BEGIN;

CREATE TABLE IF NOT EXISTS user_notification (
    id                                                  SERIAL PRIMARY KEY,
    user_id                                             uuid NOT NULL,
    notification_id                                     INT NOT NULL,
    active                                              BOOLEAN NOT NULL,
    seen_at                                             TIMESTAMP WITH TIME ZONE DEFAULT null,
    created_at                                          TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at                                          TIMESTAMP WITH TIME ZONE DEFAULT now(),
    deleted_at                                          TIMESTAMP WITH TIME ZONE DEFAULT null,
    CONSTRAINT fk_user_notification_user_id             FOREIGN KEY(user_id) REFERENCES users(user_id),
    CONSTRAINT fk_user_notification_notification_id     FOREIGN KEY(notification_id) REFERENCES notification(id)
);

COMMIT;