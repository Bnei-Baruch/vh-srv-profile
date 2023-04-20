BEGIN;

INSERT INTO notification (slug, content) VALUES 

('mb_problem_previous_payment', '{ "en": "There was a problem with the previous payment", "ru": "There was a problem with the previous payment" }'),
('mb_expiration_notice', '{ "en": "Your membership has expired due to lack of funds", "ru": "Your membership has expired due to lack of funds" }'),
('mb_cancelled', '{ "en": "Your membership was cancelled", "ru": "Your membership was cancelled" }'),
('mb_new', '{ "en": "You do not have a membership yet", "ru": "You do not have a membership yet" }'),
('hh_request_received', '{ "en": "Your request for help haver was received", "ru": "Your request for help haver was received" }'),
('hh_request_approved', '{ "en": "Your request for help haver was approved", "ru": "Your request for help haver was approved" }'),
('hh_request_refused', '{ "en": "Your request for help haver was refused", "ru": "Your request for help haver was refused" }'),
('mb_has_expired_notice', '{ "en": "Your membership has expired", "ru": "Your membership has expired" }') 
ON CONFLICT DO NOTHING;

COMMIT;