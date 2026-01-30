begin;

create table if not exists shared_data(
    player_id uuid not null,
    profile_id uuid not null,
    data jsonb,
    constraint player_profile unique (player_id, profile_id)
);

commit;