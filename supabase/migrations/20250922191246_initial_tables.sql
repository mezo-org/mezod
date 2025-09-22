create table assets_unlocked_events (
  unlock_sequence character varying not null,
  recipient character varying not null,
  token character varying not null,
  sender character varying not null,
  amount character varying not null,
  chain int not null,
  block_time numeric not null,
  constraint assets_unlocked_events_pkey primary key (unlock_sequence)
);

create table signatures (
  unlock_sequence character varying not null,
  signature character varying not null,
  constraint deposits_uc unique (signature)
 );
