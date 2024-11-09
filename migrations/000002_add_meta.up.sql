create table if not exists answer_meta (
  id serial primary key,
  message_id int not null references messages(id),
  slide_num int not null,
  file_id uuid not null,
  file_name varchar not null
)