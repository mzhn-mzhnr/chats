create table if not exists answer_meta (
  message_id int primary key not null references messages(id),
  slide_num int not null,
  file_id uuid not null,
  file_name varchar not null
)