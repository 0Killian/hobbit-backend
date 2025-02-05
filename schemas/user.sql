create table user (
	user_id uuid primary key not null default gen_random_uuid(),
	cloud_iam_sub uuid not null
);

create table user_experience (
	user_id uuid primary key not null,
	rank decimal not null,

	foreign key (user_id) references "user"(user_id)
)
