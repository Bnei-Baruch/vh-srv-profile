BEGIN;

DELETE FROM notification WHERE slug IN (
    'mb_problem_previous_payment',
    'mb_expiration_notice',
    'mb_cancelled',
    'mb_new',
    'hh_request_received',
    'hh_request_approved',
    'hh_request_refused',
    'mb_has_expired_notice'
);

COMMIT;