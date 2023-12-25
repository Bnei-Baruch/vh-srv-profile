
-- cleanup duplicate expiry notifications
select user_id, count(*), array_agg(notification_id)
from user_notification
where notification_id in (2, 8)
  and active = true
group by user_id
having count(*) > 1;

delete
from user_notification
where notification_id = 2
  and active = true
  and user_id in (select user_id
                  from user_notification
                  where notification_id in (2, 8)
                    and active = true
                  group by user_id
                  having count(*) > 1);

